package mapping

var (
	// ApplicationEventType maps which application sends which event type.
	ApplicationEventType = map[string]string{
		"commerce": "order.created",
		"appname":  "DocuSing_BO.Account_DocuSign.Updated",
		"no-app":   "New.Some-Other.Order-äöüÄÖÜβ.Final.C-r-e-a-t-e-d",
	}
)
