package subscriber

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func Handler(writer http.ResponseWriter, request *http.Request) {
	if b, err := httputil.DumpRequest(request, true); err != nil {
		fmt.Printf("failed to dump request with error:[%v]\n", err)
	} else {
		fmt.Println()
		fmt.Println(string(b))
		fmt.Println("--------------------------------------------------------------------------")
	}
	writer.WriteHeader(http.StatusOK)
}
