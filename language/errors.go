package language

type GraphQLError struct {
	Message string
}

func (err *GraphQLError) Error() string {
	return err.Message
}
