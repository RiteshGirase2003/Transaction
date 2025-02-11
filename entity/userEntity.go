package entity

// # User represents a user in the system.
//
// It contains all the details associated with a user including their unique identifier, role, contact information, address, and password.
// Fields:
//  - UserID: 		The unique identifier for the user.
// 	- Role: 		The role assigned to the user (e.g., 'admin', 'user').
// 	- Email: 		The email address associated with the user account.
// 	- Phone: 		The phone number of the user.
// 	- Address: 		The physical address of the user.
// 	- Name: 		The full name of the user.
// 	- Status: 		The current status of the user (e.g., 'active', 'inactive').
// 	- Password: 	The password used by the user to authenticate.
type User struct {
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Password string `json:"password"`
}
