package linear

// Define the structure of an issue node
type IssueNode struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"state"`
	Team struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
	Project *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
	Assignee *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"assignee"`
}

type IssuesConnection struct {
	Nodes []IssueNode `json:"nodes"`
}

// Define the structure of the issues response for listing
type IssuesResponseData struct {
	Issues IssuesConnection `json:"issues"`
}

// Define the structure of the issue creation response
type IssueCreateResponseData struct {
	IssueCreate struct {
		Success bool `json:"success"`
		Issue   struct {
			ID          string    `json:"id"`
			Title       string    `json:"title"`
			Description string    `json:"description"`
			Team        string    `json:"name"` // This seems incorrect, should probably be a Team struct
			Project     *struct { // Needs to be nullable like in IssueNode
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"project"`
			Assignee *struct { // Needs to be nullable like in IssueNode
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"assignee"`
			State struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"state"`
		} `json:"issue"`
	} `json:"issueCreate"`
}

// Define the structure for Teams
type TeamNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TeamsConnection struct {
	Nodes []TeamNode `json:"nodes"`
}

type TeamsResponseData struct {
	Teams TeamsConnection `json:"teams"`
}

// Define the structure for Projects
type ProjectNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProjectConnection struct {
	Nodes []ProjectNode `json:"nodes"`
}

type TeamProjectsResponseData struct {
	Team struct {
		ID       string            `json:"id"`
		Name     string            `json:"name"`
		Projects ProjectConnection `json:"projects"`
	} `json:"team"` // Added json tag for consistency
}

// Define the structure for Users/Members
type UserNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserConnection struct {
	Nodes []UserNode `json:"nodes"`
}

// Corrected TeamMembersResponseData struct using 'members' field
type TeamMembersResponseData struct {
	Team struct {
		ID      string         `json:"id"`
		Name    string         `json:"name"`
		Members UserConnection `json:"members"` // <-- Use 'members' to get assignable users
	} `json:"team"`
}

// Define the structure for States
type StateNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type StateConnection struct {
	Nodes []StateNode `json:"nodes"`
}

type TeamStatesResponseData struct {
	Team struct {
		ID     string          `json:"id"`
		Name   string          `json:"name"`
		States StateConnection `json:"states"`
	} `json:"team"`
}

// promptContent struct (can be here or in a utils file if used more widely)
type PromptContent struct {
	ErrorMsg string
	Label    string
}
