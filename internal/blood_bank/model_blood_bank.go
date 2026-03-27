package blood_bank

type BloodBank struct {
	Id        string     `json:"id"`
	Name      string     `json:"name"`
	BloodBags []BloodBag `json:"bloodBags,omitempty"`
}
