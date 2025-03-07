package domain

// Event types
const (
	OrderCreatedEventType    = "OrderCreated"
	PaymentApprovedEventType = "PaymentApproved"
	PaymentDeclinedEventType = "PaymentDeclined"
	ProductsReservedEventType = "ProductsReserved"
	ProductsShippedEventType = "ProductsShipped"
)

// Kafka topics
const (
	OrdersTopic = "orders"
	PaymentsTopic = "payments"
	InventoryTopic = "inventory"
	ShippingTopic = "shipping"
)
