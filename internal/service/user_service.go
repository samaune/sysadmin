package service

import "ndk/internal/model"

func GetAllUsers() []model.User {
	return []model.User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}
}
