package entity

import "time"

type Manager struct {
	ID           string
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
	TelegramNick string
	Phone        string
	AvatarURL    string
	CreatedAt    time.Time
}

type ManagerProfile struct {
	FirstName         string
	LastName          string
	Email             string
	TelegramNick      string
	Phone             string
	SubordinatesCount int
	AvatarURL         string
	JiraDisplayName   string
	JiraEmail         string
}
