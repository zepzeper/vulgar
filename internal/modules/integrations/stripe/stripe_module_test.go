package stripe

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

// =============================================================================
// client tests
// =============================================================================

func TestClientCreate(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local stripe = require("integrations.stripe")
		local client, err = stripe.client({api_key = "sk_test_xxx"})
		assert(err == nil, "client should not error: " .. tostring(err))
		assert(client ~= nil, "client should not be nil")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

func TestClientMissingKey(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local stripe = require("integrations.stripe")
		local client, err = stripe.client({})
		assert(client == nil or err ~= nil, "should error with missing api_key")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// create_customer tests
// =============================================================================

func TestCreateCustomerNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local stripe = require("integrations.stripe")
		local customer, err = stripe.create_customer(nil, {
			email = "test@example.com"
		})
		assert(customer == nil, "customer should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// get_customer tests
// =============================================================================

func TestGetCustomerNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local stripe = require("integrations.stripe")
		local customer, err = stripe.get_customer(nil, "cus_xxx")
		assert(customer == nil, "customer should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// create_charge tests
// =============================================================================

func TestCreateChargeNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local stripe = require("integrations.stripe")
		local charge, err = stripe.create_charge(nil, {
			amount = 1000,
			currency = "usd"
		})
		assert(charge == nil, "charge should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// create_payment_intent tests
// =============================================================================

func TestCreatePaymentIntentNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local stripe = require("integrations.stripe")
		local intent, err = stripe.create_payment_intent(nil, {
			amount = 1000,
			currency = "usd"
		})
		assert(intent == nil, "intent should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// create_subscription tests
// =============================================================================

func TestCreateSubscriptionNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local stripe = require("integrations.stripe")
		local sub, err = stripe.create_subscription(nil, {
			customer = "cus_xxx",
			price = "price_xxx"
		})
		assert(sub == nil, "sub should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// cancel_subscription tests
// =============================================================================

func TestCancelSubscriptionNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local stripe = require("integrations.stripe")
		local sub, err = stripe.cancel_subscription(nil, "sub_xxx")
		assert(sub == nil, "sub should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// list_invoices tests
// =============================================================================

func TestListInvoicesNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local stripe = require("integrations.stripe")
		local invoices, err = stripe.list_invoices(nil)
		assert(invoices == nil, "invoices should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// create_refund tests
// =============================================================================

func TestCreateRefundNoClient(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local stripe = require("integrations.stripe")
		local refund, err = stripe.create_refund(nil, {charge = "ch_xxx"})
		assert(refund == nil, "refund should be nil")
		assert(err ~= nil, "should error without client")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}

// =============================================================================
// verify_webhook tests
// =============================================================================

func TestVerifyWebhookNoSecret(t *testing.T) {
	L := newTestState()
	defer L.Close()

	err := L.DoString(`
		local stripe = require("integrations.stripe")
		local event, err = stripe.verify_webhook("payload", "signature", "")
		assert(event == nil or err ~= nil, "should error without secret")
	`)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
}
