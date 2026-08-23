"""LLM-judge metrics: faithfulness and answer relevancy.

Gemini 2.5 Flash at temperature 0 is the default judge (FIN-69); when no
Gemini key is configured the judge falls back to the service's configured
LLM provider and records it in the run manifest. Verdicts are cached by
input hash; disagreement between a first pass and a position-swapped
re-judge is inconclusive, and inconclusive is a hard failure — never a
silent pass.
"""

from __future__ import annotations

import asyncio
import hashlib
import json
import logging
from pathlib import Path
from typing import cast

from core.domain.exceptions import LLMError
from core.ports.llm import LLMPort, LLMResult
from evals.models import ClaimVerdict, GoldenItem, ItemTrace, JudgeVerdict, Locale

logger = logging.getLogger(__name__)

RUBRIC_VERSION = "1.0.0"
DEFAULT_SWAP_BAND = 0.15
REJUDGE_BELOW = 0.85

_FAITHFULNESS_RUBRIC: dict[str, str] = {
    "en": (
        "You are a strict factuality judge. Given CONTEXT passages and a CLAIM, decide "
        "whether the claim is fully supported by the context. Support requires every "
        "factual statement in the claim to be traceable to the context. If the context "
        "does not contain the information, the claim is not supported — do not use "
        "outside knowledge. Answer with JSON only: "
        '{"supported": true|false, "confidence": "high"|"low", "reason": "<one sentence>"}'
    ),
    "am": (
        "እርስዎ ጥብቅ የእውነታ ተቆጣጣሪ ናችሁ። የተሰጡዎትን CONTEXT አንቀጾች እና CLAIM ተከትላችሁ፣ ይህ ክሊይም "
        "ሙሉ በሙሉ በአውዱ ይደገፋል ወይም አይደለም ወስኑ። ድጋፍ ማለት ክሊይሙ ውስጥ ያለው እያንዳንዱ የእውነታ መግለጫ "
        "ከአውዱ መከተል እንዳለበት ነው። መረጃው በአውዱ ውስጥ ከሌለ ክሊይሙ አይደገፍም — ከውጭ እውቀት አይጠቀሙ። "
        "መልስዎን በJSON ብቻ ይስጡ፦ "
        '{"supported": true|false, "confidence": "high"|"low", "reason": "<በአንድ ሐረግ>"}'
    ),
}

_RELEVANCY_RUBRIC: dict[str, str] = {
    "en": (
        "You are an answer-quality judge. Rate how well the ANSWER responds to the "
        "QUESTION, using the REFERENCE ANSWER as a guide to what a complete response "
        "contains. The answer need not match the reference word-for-word; it must "
        "address what was asked without irrelevant or evasive content. "
        "Score from 0.0 (no answer to the question) to 1.0 (fully responsive). "
        "Answer with JSON only: "
        '{"score": 0.0-1.0, "confidence": "high"|"low", "reason": "<one sentence>"}'
    ),
    "am": (
        "እርስዎ የመልስ ጥራት ተቆጣጣሪ ናችሁ። ANSWER ለQUESTION ምን ያህል መልስ እንደሚሰጥ ያረጋግጡ፤ "
        "REFERENCE ANSWER ሙሉ መልስ ምን እንደሚይዝ የሚመለከት መመሪያ ነው። መልሱ ከማጣቀሻው ቃል በቃል "
        "እንዲመሳከል አያስፈልገውም፤ የተጠየቀውን ርዕስ ያለ ውድቅ ወይም መራቅ በቀጥታ እንዲይዝ ግን አለበት። "
        "ከ0.0 (ጥያቄውን አይመልስም) እስከ 1.0 (ሙሉ በሙሉ ይመልሳል) ያሸቡ። "
        "መልስዎን በJSON ብቻ ይስጡ፦ "
        '{"score": 0.0-1.0, "confidence": "high"|"low", "reason": "<በአንድ ሐረግ>"}'
    ),
}

_EXTRACT_RUBRIC = (
    "Extract the distinct factual claims made in the ANSWER. A claim is a single "
    "checkable statement (a fact, number, date, requirement, or attribution). "
    "Ignore hedging, greetings, and advice to consult professionals. Return at most "
    "8 claims, in the answer's language. Answer with JSON only: "
    '{"claims": ["...", "..."]}'
)

_UNKNOWN_RUBRIC: dict[str, str] = {
    "en": (
        "You are evaluating whether an assistant honestly declined to answer a question "
        "its sources could not support. A response counts as an honest refusal if it "
        "declines to provide the requested fact, says the information is not available "
        "in its documents/sources, or otherwise withholds a specific answer — even if "
        "worded differently from a canonical refusal. It does NOT count as a refusal if "
        "it states the requested fact or figure. Answer with JSON only: "
        '{"refused": true|false, "confidence": "high"|"low", "reason": "<one sentence>"}'
    ),
    "am": (
        "እርስዎ አርታኤው ለጥያቄው የተረጋገጠ መረጃ ሳያገኝ በትክክል መታወቂያውን መግለጹን ይገምግማሉ። መታወቂያ "
        "የሚቆጠረው አርታኤው የተጠየቀውን እውነታ ወይም ቁጥር ሳይሰጥ፣ መረጃው በሰነዶቹ ወይም ምንጮቹ ውስጥ "
        "እንደሌለ ቢል ወይም መልስ መስጠት አለመቻሉን ቢገልጽ ነው። የተጠየቀውን እውነታ ወይም ቁጥር ከሰጠ ግን "
        "መታወቂያ አይቆጠርም። መልስዎን በJSON ብቻ ይስጡ፦ "
        '{"refused": true|false, "confidence": "high"|"low", "reason": "<በአንድ ሐረግ>"}'
    ),
}


class JudgeError(RuntimeError):
    """Raised when the judge cannot produce a verdict at all."""


JUDGE_RETRY_DELAYS_SECONDS = (15.0, 30.0, 60.0)


class EvalJudge:
    def __init__(
        self,
        llm: LLMPort,
        *,
        cache_dir: Path | None = None,
        swap_band: float = DEFAULT_SWAP_BAND,
    ) -> None:
        self._llm = llm
        self._cache_dir = cache_dir
        self._swap_band = swap_band
        if cache_dir is not None:
            cache_dir.mkdir(parents=True, exist_ok=True)

    @property
    def provider(self) -> str:
        return self._llm.provider

    @property
    def model(self) -> str:
        return self._llm.model

    async def faithfulness(self, item: GoldenItem, trace: ItemTrace) -> JudgeVerdict:
        try:
            claims = await self._extract_claims(trace.answer)
        except (JudgeError, LLMError) as exc:
            return JudgeVerdict(inconclusive=True, rationale=str(exc))
        if not claims:
            return JudgeVerdict(score=1.0, rationale="no factual claims to verify")
        context = "\n\n---\n\n".join(trace.context_texts)
        if not context.strip():
            return JudgeVerdict(
                inconclusive=True,
                rationale="no judge context captured for faithfulness",
            )
        supported = 0
        inconclusive_claims = 0
        for claim in claims:
            verdict = await self._verify_claim(claim, context, item.locale)
            if verdict.inconclusive or verdict.supported is None:
                inconclusive_claims += 1
                continue
            if verdict.supported:
                supported += 1
        if inconclusive_claims:
            return JudgeVerdict(
                inconclusive=True,
                rationale=f"{inconclusive_claims}/{len(claims)} claims inconclusive",
            )
        return JudgeVerdict(
            score=supported / len(claims), rationale=f"{supported}/{len(claims)} claims supported"
        )

    async def answer_relevancy(self, item: GoldenItem, trace: ItemTrace) -> JudgeVerdict:
        try:
            first = await self._score_relevancy(item, trace, swapped=False)
            second: JudgeVerdict | None = None
            if first.score is None or first.score < REJUDGE_BELOW:
                second = await self._score_relevancy(item, trace, swapped=True)
        except (JudgeError, LLMError) as exc:
            return JudgeVerdict(inconclusive=True, rationale=str(exc))
        return self._reconcile_relevancy(first, second)

    def _reconcile_relevancy(
        self, first: JudgeVerdict, second: JudgeVerdict | None
    ) -> JudgeVerdict:
        if first.score is None:
            if second is None or second.score is None:
                return JudgeVerdict(
                    inconclusive=True, rationale="relevancy judge unparseable twice"
                )
            return second
        if first.score >= REJUDGE_BELOW or second is None or second.score is None:
            return first
        if abs(first.score - second.score) > self._swap_band:
            return JudgeVerdict(
                inconclusive=True,
                rationale=f"relevancy judges disagree: {first.score:.2f} vs {second.score:.2f}",
                swapped_used=True,
            )
        return JudgeVerdict(
            score=round((first.score + second.score) / 2, 4),
            rationale=first.rationale,
            swapped_used=True,
        )

    async def honest_refusal(self, item: GoldenItem, answer: str) -> JudgeVerdict:
        rubric = _UNKNOWN_RUBRIC[item.locale.value]
        try:
            payload = await self._cached_json(
                "unknown_refusal",
                {"query": item.query, "answer": answer, "locale": item.locale.value},
                rubric,
                f"QUESTION: {item.query}\n\nANSWER: {answer}",
            )
        except (JudgeError, LLMError) as exc:
            return JudgeVerdict(inconclusive=True, rationale=str(exc))
        refused = payload.get("refused")
        if not isinstance(refused, bool):
            return JudgeVerdict(inconclusive=True, rationale="unknown-refusal verdict unparseable")
        return JudgeVerdict(
            score=1.0 if refused else 0.0,
            rationale=str(payload.get("reason", "")),
        )

    async def _extract_claims(self, answer: str) -> list[str]:
        payload = await self._cached_json(
            "claims",
            {"answer": answer},
            _EXTRACT_RUBRIC,
            answer,
        )
        raw_claims = payload.get("claims")
        if not isinstance(raw_claims, list):
            return []
        claims: list[str] = [
            claim
            for claim in cast(list[object], raw_claims)
            if isinstance(claim, str) and claim.strip()
        ]
        return claims[:8]

    async def _verify_claim(
        self,
        claim: str,
        context: str,
        locale: Locale,
    ) -> ClaimVerdict:
        rubric = _FAITHFULNESS_RUBRIC[locale.value]
        payload = await self._cached_json(
            "claim_support",
            {"context": context, "claim": claim, "locale": locale.value},
            rubric,
            f"CONTEXT:\n{context}\n\nCLAIM: {claim}",
        )
        supported = payload.get("supported")
        confidence = payload.get("confidence", "high")
        if not isinstance(supported, bool):
            return ClaimVerdict(claim=claim, inconclusive=True, rationale="unparseable verdict")
        if confidence == "low":
            swapped_payload = await self._cached_json(
                "claim_support_swapped",
                {"context": context, "claim": claim, "locale": locale.value},
                rubric,
                f"CLAIM: {claim}\n\nCONTEXT:\n{context}",
            )
            swapped_supported = swapped_payload.get("supported")
            if not isinstance(swapped_supported, bool) or swapped_supported != supported:
                return ClaimVerdict(
                    claim=claim,
                    inconclusive=True,
                    rationale="low-confidence verdict not confirmed on swap",
                )
        return ClaimVerdict(
            claim=claim, supported=supported, rationale=str(payload.get("reason", ""))
        )

    async def _score_relevancy(
        self,
        item: GoldenItem,
        trace: ItemTrace,
        *,
        swapped: bool,
    ) -> JudgeVerdict:
        rubric = _RELEVANCY_RUBRIC[item.locale.value]
        if swapped:
            user_content = (
                f"REFERENCE ANSWER: {item.golden_answer}\n\n"
                f"QUESTION: {item.query}\n\nANSWER: {trace.answer}"
            )
            cache_key = "relevancy_swapped"
        else:
            user_content = (
                f"QUESTION: {item.query}\n\n"
                f"REFERENCE ANSWER: {item.golden_answer}\n\nANSWER: {trace.answer}"
            )
            cache_key = "relevancy"
        payload = await self._cached_json(
            cache_key,
            {
                "query": item.query,
                "golden": item.golden_answer,
                "answer": trace.answer,
                "locale": item.locale.value,
            },
            rubric,
            user_content,
        )
        score = payload.get("score")
        if isinstance(score, bool) or not isinstance(score, (int, float)):
            return JudgeVerdict(inconclusive=True, rationale="relevancy score unparseable")
        clamped = min(1.0, max(0.0, float(score)))
        return JudgeVerdict(
            score=clamped,
            rationale=str(payload.get("reason", "")),
            swapped_used=swapped,
        )

    async def _cached_json(
        self,
        metric: str,
        inputs: dict[str, str],
        system_rubric: str,
        user_content: str,
    ) -> dict[str, object]:
        cache_path: Path | None = None
        if self._cache_dir is not None:
            digest = hashlib.sha256(
                json.dumps(
                    {
                        "metric": metric,
                        "rubric_version": RUBRIC_VERSION,
                        "model": f"{self._llm.provider}/{self._llm.model}",
                        "inputs": inputs,
                    },
                    sort_keys=True,
                    ensure_ascii=False,
                ).encode("utf-8")
            ).hexdigest()
            cache_path = self._cache_dir / f"{metric}-{digest}.json"
            if cache_path.exists():
                try:
                    cached = json.loads(cache_path.read_text(encoding="utf-8"))
                    if isinstance(cached, dict):
                        return cast(dict[str, object], cached)
                except (OSError, ValueError):
                    logger.warning("judge cache unreadable, recomputing: %s", cache_path)
        raw = await self._generate_json(system_rubric, user_content)
        if cache_path is not None:
            cache_path.write_text(json.dumps(raw, ensure_ascii=False), encoding="utf-8")
        return raw

    async def _generate_json(self, system_rubric: str, user_content: str) -> dict[str, object]:
        result: LLMResult | None = None
        last_error: LLMError | None = None
        for attempt, delay in enumerate((0.0, *JUDGE_RETRY_DELAYS_SECONDS), start=1):
            if delay:
                logger.warning(
                    "judge provider error, retry %d/%d in %.0fs",
                    attempt - 1,
                    len(JUDGE_RETRY_DELAYS_SECONDS),
                    delay,
                )
                await asyncio.sleep(delay)
            try:
                result = await self._llm.generate(
                    user_content,
                    system_prompt=system_rubric,
                    max_tokens=1024,
                    temperature=0.0,
                )
                break
            except LLMError as exc:
                last_error = exc
        if result is None:
            msg = f"judge provider failed after retries: {last_error}"
            raise JudgeError(msg)
        parsed = _extract_json_object(result.text)
        if parsed is None:
            msg = f"judge returned unparseable JSON: {result.text[:200]!r}"
            raise JudgeError(msg)
        return parsed


def _extract_json_object(text: str) -> dict[str, object] | None:
    cleaned = text.strip()
    if cleaned.startswith("```"):
        cleaned = cleaned.removeprefix("```json").removeprefix("```")
        cleaned = cleaned.split("```", maxsplit=1)[0]
    start = cleaned.find("{")
    end = cleaned.rfind("}")
    if start == -1 or end <= start:
        return None
    try:
        parsed = json.loads(cleaned[start : end + 1])
    except ValueError:
        return None
    return cast(dict[str, object], parsed) if isinstance(parsed, dict) else None


__all__ = ["RUBRIC_VERSION", "EvalJudge", "JudgeError", "_extract_json_object"]
