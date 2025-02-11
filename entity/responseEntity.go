package entity

// # CommonResponse represents the standard structure for API response messages.
//
// It is used to standardize the structure of responses sent from the API, including a status code and a message.
//
// 	- Status: 	A numeric value representing the result of the API request (e.g., 200 for success, 400 for failure).
// 	- Message: 	A descriptive message that provides more information about the response status (e.g., "API Success", "API Failure").
type CommonResponse struct {
	Status  int    `json:"status"`
	Message string `json:"msg"`
}

// # StatusName represents the type used for status names in API responses.
//
// It is an integer type that maps to specific status codes (e.g., SUCCESS, FAILURE).
// The constants defined for this type can be used to identify different response statuses.
type StatusName int

// # StatusInfo stores detailed information about a specific status.
//
// It is used to associate a status name with an HTTP status code and a descriptive message.
//	 - Status: 		The HTTP status code associated with this status (e.g., 200 for success, 400 for failure).
//	 - Message: 	A string message providing context for the status (e.g., "API Success", "API Failure").
type StatusInfo struct {
	Status  int
	Message string
}

// # Constants representing the status names used throughout the API responses.
//
// These constants can be used to refer to specific statuses in the response and to set the corresponding HTTP status code and message.
// 	- SUCCESS: 					Indicates that the API request was successful (status code 200).
// 	- FAILURE: 					Indicates that the API request failed (status code 400).
// 	- COMMON_SERVER_ERROR: 		Indicates that there was an internal server error (status code 500).
const (
	SUCCESS StatusName = iota
	FAILURE
	COMMON_SERVER_ERROR
)

// # StatusEnum is a map that associates each StatusName constant with its corresponding StatusInfo.
//
// It holds the status codes and messages for different response statuses.
// This map is used to easily fetch the status code and message for a given status name (e.g., SUCCESS, FAILURE).
//
// will give you the StatusInfo for a successful request (status code 200, "API Success").
var StatusEnum = map[StatusName]StatusInfo{
	SUCCESS: {
		Status:  200,
		Message: "API Success",
	},
	FAILURE: {
		Status:  400,
		Message: "API Failure",
	},
	COMMON_SERVER_ERROR: {
		Status:  500,
		Message: "Error Occurred in internal server",
	},
}

// ApplyResponseBody is a method that updates the Status and Message fields of the CommonResponse struct based on the provided status name.
//
// It retrieves the status code and message from the StatusEnum map and sets them in the response.
//
// 	- statusName: The status name (e.g., SUCCESS, FAILURE) that determines the response status and message.
// This method is typically used to standardize the API response format based on the outcome of an API request.
func (commonResponse *CommonResponse) ApplyResponseBody(statusName StatusName) {
	commonResponse.Status = StatusEnum[statusName].Status
	commonResponse.Message = StatusEnum[statusName].Message
}
