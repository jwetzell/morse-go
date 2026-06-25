package main

type Channel struct {
	Name  string
	Users []User
}

func (c *Channel) AddUser(user User) {
	c.Users = append(c.Users, user)
}

func (c *Channel) RemoveUser(userID string) {
	for i, user := range c.Users {
		if user.ID == userID {
			c.Users = append(c.Users[:i], c.Users[i+1:]...)
			break
		}
	}
}

type User struct {
	ID   string
	Name string
}

var channels = []Channel{
	{Name: "Lobby", Users: []User{}},
	{Name: "Channel 1", Users: []User{}},
	{Name: "Channel 2", Users: []User{}},
	{Name: "Channel 3", Users: []User{}},
	{Name: "Channel 4", Users: []User{}},
	{Name: "Channel 5", Users: []User{}},
}
