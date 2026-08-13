from __future__ import annotations

from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict

AI_PERSONA_SYSTEM_PROMPT: str = """
You are the "Adisu Serategna AI Advisor", an expert, empathetic, and highly precise Ethiopian Business and Regulatory Assistant.
Your goal is to help Ethiopian Micro, Small, and Medium Enterprises (MSMEs) navigate bureaucracy, formalize their businesses, and achieve compliance.
You speak both English and Amharic fluently. Always respond in the language the user used to ask the question.
Your tone is encouraging but strictly professional. You are an expert guide, but you MUST state that you do not provide certified legal or financial counsel.
"""

AI_RESTRICTIONS: str = """
When answering the user's query, you MUST adhere to the following strict directives:

1. HYPER-PERSONALIZATION:
Always tailor your advice to the USER BUSINESS PROFILE. If they are in the OROMIA region, prioritize Oromia regional directives over Addis Ababa directives. If they have the `tax-tot` tag, do NOT explain VAT rules unless explicitly asked.
2. GROUNDED TRUTH (RAG INVOCATION):
For any question regarding taxes, laws, penalties, or government procedures, you MUST use the `search_knowledge_base` tool. DO NOT hallucinate Ethiopian laws or rely on your pre-trained internet knowledge.
3. CITATION MANDATE:
If you provide regulatory information, you must cite the specific document or proclamation returned by your knowledge base tool. (e.g., "According to the Ethiopian Income Tax Proclamation No. 979/2016...")
4. ACTION-ORIENTED SUPPORT:
You are part of the Adisu Serategna platform. If the user needs a document, use the `find_template` tool to see if we have it in our library. If the user asks about a registration step, use the `check_guide_progress` tool to see where they are in their formalization journey.
5. HANDLING UNKNOWNS:
If your tools do not return relevant Ethiopian context, honestly state: "I currently do not have the verified Ethiopian regulatory documents to answer this specific question." Do not guess.
"""


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
    )

    APP_NAME: str = "ai-service"
    APP_VERSION: str = "0.1.0"
    DEBUG: bool = False
    LOG_LEVEL: str = "INFO"

    HTTP_PORT: int = 8000
    CORS_ORIGINS: list[str] = ["http://localhost:3000"]
    # When True, httpx honors HTTP(S)_PROXY from the environment (default httpx behavior).
    HTTPX_TRUST_ENV: bool = True

    GRPC_PORT: int = 50051
    CORE_GRPC_ENDPOINT: str = "localhost:50052"
    INTERNAL_GRPC_AUTH_TOKEN: str = ""

    SEAWEEDFS_FILER_URL: str = "http://localhost:8888"

    DATABASE_URL: str = "postgresql+asyncpg://postgres:postgres@localhost:5432/adisu_ai"

    REDIS_URL: str = "redis://localhost:6379/0"

    RABBITMQ_URL: str = "amqp://guest:guest@localhost:5672/"

    INGESTION_WORKER_QUEUE: str = "ai.ingestion.requested.v1"
    INGESTION_WORKER_EXCHANGE: str = "core.events"
    INGESTION_WORKER_ROUTING_KEY: str = "document.ingestion.requested.v1"
    INGESTION_WORKER_PREFETCH_COUNT: int = 16
    INGESTION_WORKER_REQUEUE_ON_FAILURE: bool = True
    INGESTION_WORKER_MAX_RETRIES: int = 3
    INGESTION_WORKER_RETRY_BASE_DELAY_MS: int = 1000
    INGESTION_WORKER_RETRY_MAX_DELAY_MS: int = 30000
    INGESTION_WORKER_RETRY_JITTER_FACTOR: float = 0.1

    INGESTION_WORKER_DLQ_ENABLED: bool = True
    INGESTION_WORKER_DLQ_EXCHANGE: str = "ai.ingestion.dlq"
    INGESTION_WORKER_DLQ_ROUTING_KEY: str = "document.ingestion.dlq.v1"

    # Ingestion envelope verification
    INGESTION_SIGNING_ACTIVE_KEY_ID: str = "ingestion-v1"
    INGESTION_SIGNING_ACTIVE_KEY_SECRET: str = "change-me"
    INGESTION_SIGNING_PREVIOUS_KEYS_JSON: str = "{}"

    # Embeddings
    EMBEDDING_PROVIDER: str = "cohere"
    COHERE_API_KEY: str = ""
    GEMINI_API_KEY: str = ""
    COHERE_EMBEDDING_MODEL: str = "embed-multilingual-v3.0"
    GEMINI_EMBEDDING_MODEL: str = "text-multilingual-embedding-002"
    OLLAMA_EMBEDDING_MODEL: str = "nomic-embed-text"
    OLLAMA_BASE_URL: str = "http://localhost:11434"
    EMBEDDING_DIMENSIONS: int | None = None

    LLM_PROVIDER: str = "cohere"
    COHERE_LLM_MODEL: str = "command-a-03-2025"
    GEMINI_LLM_MODEL: str = "gemini-2.5-flash"
    OLLAMA_LLM_MODEL: str = "qwen2.5"
    AI_ASK_ENABLED: bool = True

    # Gemini / Vertex AI
    GEMINI_USE_VERTEX: bool = True
    GEMINI_VERTEX_PROJECT: str = ""
    GEMINI_VERTEX_LOCATION: str = "us-central1"

    AI_PROMPT_DIR: str = "prompts"
    AI_AGENTIC_ENABLED: bool = True
    AI_AGENTIC_MAX_ITERATIONS: int = 5

    # Agentic KB follow-up query policy (FIN-75): duplicate/drift guards.
    AI_KB_DUP_COSINE_THRESHOLD: float = 0.92
    AI_KB_DUP_AMBIGUOUS_BAND_MIN: float = 0.80
    AI_KB_DUP_JACCARD_THRESHOLD: float = 0.6
    AI_KB_DRIFT_COSINE_THRESHOLD: float = 0.40
    AI_KB_SUPPRESSION_TRIPWIRE: int = 2

    AI_EXPECTED_REMOTE_TOOLS: list[str] = [
        "search_guides",
        "get_user_profile",
        "find_template",
        "get_guide_progress",
        "check_compliance_status",
        "list_sectors",
        "list_tags",
        "get_guide_detail",
    ]
    AI_TOOL_REFRESH_TTL_SEC: int = 300

    AI_INTENT_SIMILARITY_THRESHOLD: float = 0.6

    AI_INTENT_SEED_QUERIES_KNOWLEDGE: list[str] = [
        "business registration process",
        "how to register a business in ethiopia",
        "tax rates for small businesses",
        "license types for restaurants",
        "regulation for importers",
        "trade license requirements",
        "business formalization steps",
        "what permits do I need",
        "renewal process for trade license",
        "tax filing deadlines ethiopia",
        "vat registration requirements",
        "business name registration",
        "penalties for late filing",
        "investment incentives ethiopia",
        "labor law requirements",
        "environmental impact assessment",
        "construction permit process",
        "food handling certificate",
        "export license requirements",
        "professional license renewal",
    ]

    AI_INTENT_SEED_QUERIES_PERSONAL: list[str] = [
        "my business profile",
        "what is my current status",
        "show my compliance deadlines",
        "my guide progress",
        "where am I stuck",
        "what step am I on",
        "my account settings",
        "show my reminders",
        "my upcoming deadlines",
        "check my compliance status",
        "am I compliant",
        "what have I completed so far",
        "my recent activity",
        "show my dashboard",
        "my business information",
        "what is my sector",
        "my registration stage",
        "how far along am I",
        "what is missing in my profile",
        "my pending tasks",
    ]

    OTEL_ENABLED: bool = False
    PROMETHEUS_PORT: int = 9090


@lru_cache
def get_settings() -> Settings:
    return Settings()
