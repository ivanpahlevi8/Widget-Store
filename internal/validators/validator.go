package validators

import "log"

// create validators object
type Validators struct {
	Error map[string]string
}

// create function to create object
func New() *Validators {
	validate := Validators{Error: make(map[string]string)}

	return &validate
}

// cretae function to validate is error exist or not
func (valid *Validators) Validate() bool {
	// create variable bool
	var result bool

	if len(valid.Error) == 0 {
		result = true
	} else {
		result = false
	}

	log.Println("result validate : ", result)

	return result
}

// create function to add error to validators
func (valid *Validators) AddErr(key, message string) {
	// check it error already exist or not by checking value
	_, value := valid.Error[key]

	// check if value exist
	if !value {
		valid.Error[key] = message
	}
}

// cretae function to check
func (valid *Validators) Check(ok bool, key, message string) {
	if !ok {
		log.Println("adding error")
		valid.Error[key] = message
	}
}
