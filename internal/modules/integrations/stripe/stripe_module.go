package stripe

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.stripe"

// luaConfigure configures the Stripe client
// Usage: local client, err = stripe.configure({api_key = "sk_..."})
func luaConfigure(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCreateCustomer creates a customer
// Usage: local customer, err = stripe.create_customer(client, {email = "user@example.com", name = "John"})
func luaCreateCustomer(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaGetCustomer gets a customer by ID
// Usage: local customer, err = stripe.get_customer(client, customer_id)
func luaGetCustomer(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaListCustomers lists customers
// Usage: local customers, err = stripe.list_customers(client, {limit = 10})
func luaListCustomers(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCreatePaymentIntent creates a payment intent
// Usage: local intent, err = stripe.create_payment_intent(client, {amount = 1000, currency = "usd"})
func luaCreatePaymentIntent(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaConfirmPaymentIntent confirms a payment intent
// Usage: local intent, err = stripe.confirm_payment_intent(client, intent_id, {payment_method = "pm_..."})
func luaConfirmPaymentIntent(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCreateCharge creates a charge (legacy)
// Usage: local charge, err = stripe.create_charge(client, {amount = 1000, currency = "usd", source = "tok_..."})
func luaCreateCharge(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCreateRefund creates a refund
// Usage: local refund, err = stripe.create_refund(client, {charge = "ch_...", amount = 500})
func luaCreateRefund(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCreateSubscription creates a subscription
// Usage: local subscription, err = stripe.create_subscription(client, {customer = "cus_...", items = {{price = "price_..."}}})
func luaCreateSubscription(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCancelSubscription cancels a subscription
// Usage: local subscription, err = stripe.cancel_subscription(client, subscription_id)
func luaCancelSubscription(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCreateInvoice creates an invoice
// Usage: local invoice, err = stripe.create_invoice(client, {customer = "cus_..."})
func luaCreateInvoice(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaVerifyWebhookSignature verifies webhook signature
// Usage: local event, err = stripe.verify_webhook(payload, signature, webhook_secret)
func luaVerifyWebhookSignature(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

var exports = map[string]lua.LGFunction{
	"configure":              luaConfigure,
	"create_customer":        luaCreateCustomer,
	"get_customer":           luaGetCustomer,
	"list_customers":         luaListCustomers,
	"create_payment_intent":  luaCreatePaymentIntent,
	"confirm_payment_intent": luaConfirmPaymentIntent,
	"create_charge":          luaCreateCharge,
	"create_refund":          luaCreateRefund,
	"create_subscription":    luaCreateSubscription,
	"cancel_subscription":    luaCancelSubscription,
	"create_invoice":         luaCreateInvoice,
	"verify_webhook":         luaVerifyWebhookSignature,
}

// Loader is called when the module is required via require("stripe")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
