
package entity

// # Login represents the structure for user login credentials.
//
// This struct is used to capture the user's login information, such as their unique user ID, email, and password.
//
// Fields:
//
// 	1. UserID:		 A unique identifier for the user. It is a string field that holds the user's ID.
// 	2. Email: 		 The email address associated with the user's account. This is a string field.
// 	3. Password: 	 The password for the user to authenticate. This field is required and is validated using the "binding" tag.
//
// 
// 	- The `binding:"required"` tag ensures that the password field must not be empty when processing the login request.
//
type Login struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
}
