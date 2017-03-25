package main

import "testing"

// func TestAddUser(t *testing.T) {
// 	u1 := &UserData{
// 		Username: "sausheong",
// 		Email:    "sausheong@gmail.com",
// 		Name:     "Chang Sau Sheong",
// 	}
// 	err := u1.Create()
// 	if err != nil {
// 		t.Error("Cannot create", err)
// 	}
// }

func TestGetUsername(t *testing.T) {
	u := &UserData{
		Username: "sausheong",
	}

	err := u.GetByUsername()
	if err != nil {
		t.Error("Cannot get by username:", err)
	}
}
