# Flutter — Chapa Payment Integration Design

## Dependencies

Add to `mobile/pubspec.yaml`:

```yaml
dependencies:
  chapasdk: ^0.0.8+1    # Official Chapa Flutter SDK
```

Run: `flutter pub get`

## Integration Pattern

The mobile app integrates with Chapa at **two levels**:
1. **SDK level** — Chapa Flutter SDK handles the checkout UI
2. **API level** — Direct HTTP calls to the backend for payment initiation, verification, and subscription status

The app uses **retrofit** (via `api_client` package) for all backend API calls.

## Flow

### Step 1: Fetch Available Plans

**When:** On the subscription/pricing screen (no auth needed)

```dart
// Generated API client call
final response = await client.getPaymentPlans();
// Returns list of PlanDto { id, name, period, amount, currency, isActive }
```

Display as a comparison table with Basic (free) vs Pro (monthly/yearly).

### Step 2: Initiate Payment

**When:** User taps "Subscribe" on a plan

```dart
// Our custom API call (not part of auto-generated client)
final response = await client.postPaymentInitiate(
  body: PaymentInitiateRequest(
    planName: 'Pro',
    period: 'monthly',
  ),
);
// Returns { txRef, checkoutUrl, amount, currency, planName, period, expiresAt }
```

Store `txRef` locally and open Chapa checkout:

```dart
Chapa.paymentParameters(
  context: context,
  publicKey: AppConfig.chapaPublicKey,  // CHAPUBK-xxx from .env
  currency: response.currency,
  amount: response.amount.toString(),   // "19900" (minor unit) or formatted "199.00"
  email: loggedInUser.email,
  phone: loggedInUser.phone,
  firstName: loggedInUser.firstName,
  lastName: loggedInUser.lastName,
  txRef: response.txRef,
  title: 'Adisu Pro Subscription',
  desc: 'Monthly Pro plan',
  nativeCheckout: true,
  namedRouteFallBack: '/payment-result',
  onPaymentFinished: (message, reference, paidAmount) {
    // Called when user returns from Chapa
    Navigator.popUntil(context, (route) => route.isFirst);
    // Navigate to result page and poll for confirmation
  },
);
```

### Step 3: Handle Return from Chapa

**Two mechanisms run in parallel:**

#### A. Deep Link / Named Route

`namedRouteFallBack: '/payment-result'` causes Chapa to redirect back to the app after payment.

On the result page:
```dart
// Immediately poll backend for payment status
final status = await client.postPaymentVerify(
  body: PaymentVerifyRequest(txRef: txRef),
);

if (status.data.status == 'success') {
  // Show success, refresh subscription state
  ref.invalidate(subscriptionProvider);
} else if (status.data.status == 'pending') {
  // Still processing — poll with exponential backoff
  startPolling(txRef);
} else {
  // Failed
  showError();
}
```

#### B. Background Polling (fallback if app was killed)

When the app resumes or restarts, check if there's a pending `txRef`:

```dart
@override
void initState() {
  super.initState();
  final pendingTxRef = getPendingTxRef(); // from shared_preferences or local DB
  if (pendingTxRef != null) {
    verifyPayment(pendingTxRef);
  }
}
```

### Step 4: Subscription State Management

**Provider (Riverpod):**

```dart
@riverpod
class SubscriptionNotifier extends _$SubscriptionNotifier {
  @override
  Future<SubscriptionDto?> build() async {
    final apiClient = ref.read(apiClientProvider);
    final response = await apiClient.getMeSubscription();
    return response.data;
  }

  Future<void> refresh() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() => ref.read(apiClientProvider).getMeSubscription());
  }
}
```

**State:**

```dart
@freezed
class SubscriptionDto with _$SubscriptionDto {
  const factory SubscriptionDto({
    required String id,
    required String planName,
    required String planPeriod,
    required int amount,
    required String currency,
    required String status,
    required DateTime currentPeriodStart,
    required DateTime currentPeriodEnd,
    required int renewalCount,
  }) = _SubscriptionDto;

  factory SubscriptionDto.fromJson(Map<String, dynamic> json) =>
      _$SubscriptionDtoFromJson(json);
}
```

### Step 5: Entitlement Gating

In the app, check subscription before allowing access to premium features:

```dart
final subscription = ref.watch(subscriptionNotifierProvider);

subscription.when(
  data: (sub) {
    if (sub != null && sub.status == 'active' && sub.planName == 'Pro') {
      return PremiumContentPage();
    }
    return PaywallPage();
  },
  loading: () => const LoadingSpinner(),
  error: (_, __) => const ErrorPage(),
);
```

### Step 6: Renewal UI

When `currentPeriodEnd` is within 3 days:
```dart
final daysLeft = subscription.currentPeriodEnd.difference(DateTime.now()).inDays;
if (daysLeft <= 3 && subscription.status == 'active') {
  showRenewalBanner();
}
```

## Environment Config

Add to `assets/.env` files:
```
CHAPA_PUBLIC_KEY=CHAPUBK-test_xxxxxxxxxxxxxxxx
API_BASE_URL=http://10.0.2.2:3000
```

## API Client Changes

New endpoints to add to `api_client/openapi/openapi.json`:

```yaml
/payments/plans:
  get:
    tags: [Payment]
    summary: List available plans
    operationId: getPaymentPlans
    responses:
      '200': ...

/payments/initiate:
  post:
    tags: [Payment]
    summary: Initiate payment
    operationId: postPaymentInitiate
    security: [BearerAuth]
    requestBody: ...
    responses:
      '201': ...

/payments/verify:
  post:
    tags: [Payment]
    summary: Verify payment
    operationId: postPaymentVerify
    security: [BearerAuth]
    requestBody: ...
    responses:
      '200': ...

/me/subscription:
  get:
    tags: [Payment]
    summary: Get current subscription
    operationId: getMeSubscription
    security: [BearerAuth]
    responses:
      '200': ...
```

After updating the spec, regenerate:
```bash
cd api_client
dart run openapi_retrofit_generator
dart run build_runner build -d --delete-conflicting-outputs
```

Then update `api_client/lib/api_client.dart` to export new models and clients.

## WebView Consideration

For the native checkout, Chapa SDK renders wallet options directly. For web checkout fallback (future USD support), use an in-app WebView pointing to `checkoutUrl`. Listen for navigation events to detect redirect back from Chapa.