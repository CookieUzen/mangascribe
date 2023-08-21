package DB

import (
	"github.com/CookieUzen/mangascribe/Models"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang/glog"
	"fmt"
	"time"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil 
}

func VerifyPassword(hashedPassword, password string) (bool, error) {
	// Compare the hashed password with the password
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, err
	}

	return true, nil
}

// Checks if a password is valid
func ValidatePassword(password string) error {
	// If password is too long
	if len(password) > 72 {
		err := fmt.Errorf("Password is too long (maximum 72 characters)")
		glog.Info(err)
		return err
	}

	// If password is too short
	if len(password) < 8 {
		err := fmt.Errorf("Password is too short (minimum 8 characters)")
		glog.Info(err)
		return err
	}

	// TODO add strict password requirements

	return nil
}

// Checks if the username is taken
// Returns an error if there was an error checking the database
// Returns true if the username is taken, false if it is not 
func (dbm *DBManager) IsUsernameTaken(username string) (bool, error) {
	var count int64
	if err := dbm.DB.Model(&Models.Account{}).Where("username = ?", username).Count(&count).Error; err != nil {
		// Handle potential database errors if needed
		err = fmt.Errorf("Error checking if username is taken: %v", err)
		glog.Error(err)
		return false, err
	}

	// If the username is taken
	if count > 0 {
		return true, nil
	}

	return false, nil
}


// Check if an email is taken
// Returns an error if there was an error checking the database
// Returns true if the email is taken, false if it is not
func (dbm *DBManager) IsEmailTaken(email string) (bool, error) {
	var count int64
	if err := dbm.DB.Model(&Models.Account{}).Where("email = ?", email).Count(&count).Error; err != nil {
		// Handle potential database errors if needed
		err = fmt.Errorf("Error checking if email is taken: %v", err)
		glog.Error(err)
		return false, err
	}

	// If the email is taken
	if count > 0 {
		return true, nil
	}

	return false, nil
}

// Creates a new account in the database
// Returns an error if there was an error creating the account
func (dbm *DBManager) CreateAccount(account Models.NewAccountRequest) (*Models.Account, error) {
	hashedPassword, err := HashPassword(account.Password)
	if err != nil {
		return nil, err
	}

	// Check if username/email field is missing
	if account.Username == "" || account.Email == "" {
		err = fmt.Errorf("Username or email field is missing")
		glog.Error(err)
		return nil, err
	}

	// Check if the password is valid
	if err := ValidatePassword(account.Password); err != nil {
		return nil, err
	}

	// Check if username is taken
	if taken, err := dbm.IsUsernameTaken(account.Username); err != nil {
		return nil, err
	} else if taken {
		err = fmt.Errorf("Username is taken")
		glog.Info(err)
		return nil, err
	}

	// Check if email is taken
	if taken, err := dbm.IsEmailTaken(account.Email); err != nil {
		return nil, err
	} else if taken {
		err = fmt.Errorf("Email is taken")
		glog.Info(err)
		return nil, err
	}

	newAccount := Models.Account{
		Username: account.Username,
		Password: hashedPassword,
		Email: account.Email,
		API_Keys: []Models.APIKey{},
	}

	// Create the account
	if err := dbm.DB.Create(&newAccount).Error; err != nil {
		err = fmt.Errorf("Error creating account: %v", err)
		glog.Error(err)
		return nil, err
	}

	return &newAccount, nil
}

// Get an account from the database
func (dbm *DBManager) GetAccount(account *Models.Account, identifier string) error {
	if err := dbm.DB.Where("username = ?", identifier).Or("email = ?", identifier).First(&account).Error; err != nil {
		err = fmt.Errorf("Error getting account: %v", err)
		glog.Error(err)
		return err
	}

	return nil
}

// Get an account and check if the password is correct
// Returns true if the password is correct, false if it is not
func (dbm *DBManager) AuthAccount(account *Models.Account, login Models.LoginRequest) (bool, error) {
	var newAccount Models.Account
	if err := dbm.GetAccount(&newAccount, login.Identifier); err != nil {
		return false, err
	}

	if ok, err := VerifyPassword(newAccount.Password, login.Password); err != nil {
		return false, err
	} else if !ok {	// Password is incorrect
		return false, nil
	}

	*account = newAccount
	return true, nil
}


// Change the password of an account
// Returns the new hash if the password was changed successfully
func (dbm *DBManager) ChangePassword(account *Models.Account, newPassword string) (string, error) {
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return "", err
	}

	// Check if the password is valid
	if err := ValidatePassword(newPassword); err != nil {
		return "", err
	}

	// Change the password
	if err := dbm.DB.Model(account).Update("password", hashedPassword).Error; err != nil {
		err = fmt.Errorf("Error changing password: %v", err)
		glog.Error(err)
		return "", err
	}

	return hashedPassword, nil
}

// Delete an account from the database
func (dbm *DBManager) DeleteAccount(account *Models.Account) error {
	if err := dbm.DB.Delete(&account).Error; err != nil {
		err = fmt.Errorf("Error deleting account: %v", err)
		glog.Error(err)
		return err
	}

	return nil
}

// Change the email of an account
func (dbm *DBManager) ChangeEmail(account *Models.Account, newEmail string) error {
	// Check if the email is taken
	if taken, err := dbm.IsEmailTaken(newEmail); err != nil {
		return err
	} else if taken {
		err = fmt.Errorf("Email is taken")
		glog.Info(err)
		return err
	}

	// Change the email
	if err := dbm.DB.Model(account).Update("email", newEmail).Error; err != nil {
		err = fmt.Errorf("Error changing email: %v", err)
		glog.Error(err)
		return err
	}

	return nil
}

// Change the username of an account
func (dbm *DBManager) ChangeUsername(account *Models.Account, newUsername string) error {
	// Check if the username is taken
	if taken, err := dbm.IsUsernameTaken(newUsername); err != nil {
		return err
	} else if taken {
		err = fmt.Errorf("Username is taken")
		glog.Info(err)
		return err
	}

	// Change the username
	if err := dbm.DB.Model(account).Update("username", newUsername).Error; err != nil {
		err = fmt.Errorf("Error changing username: %v", err)
		glog.Error(err)
		return err
	}

	return nil
}

// Generate an API key to an account
// Returns the API key if it was generated successfully
func (dbm *DBManager) GenerateAPIKey(account *Models.Account, Duration time.Duration) (*Models.APIKey, error) {
	// Generate the API key
	key, err := account.GenerateAPIKey(Duration)
	if err != nil {
		return nil, err
	}

	// Add the API key to the database
	if err := dbm.DB.Model(account).Association("API_Keys").Append(key); err != nil {
		err = fmt.Errorf("Error adding API key: %v", err)
		glog.Error(err)
		return nil, err
	}

	return key, nil
}

// Check if an API key is valid
// returns the user associated with the API key if it is valid
// returns an error if the API key is invalid
func (dbm *DBManager) UserFromKey(account *Models.Account, key string) error {
	
	// Search db for API key
	var apiKey Models.APIKey
	if err := dbm.DB.Where("key = ?", key).First(&apiKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = fmt.Errorf("API key not found")
			glog.Info(err)
			return err
		}

		err = fmt.Errorf("Error searching db for API key: %v", err)
		glog.Error(err)
		return err
	}

	// Check if API key is expired
	if apiKey.IsExpired() {
		err := fmt.Errorf("API key is expired")
		glog.Info(err)
		return err
	}

	// Get the user associated with the API key
	if err := dbm.DB.First(&account, apiKey.AccountID).Error; err != nil {
		err = fmt.Errorf("Error getting user associated with API key: %v", err)
		glog.Error(err)
		return err
	}
	// if err := dbm.DB.Model(&apiKey).Association("ID").Find(&account); err != nil {
	// 	err = fmt.Errorf("Error getting user associated with API key: %v", err)
	// 	glog.Error(err)
	// 	return err
	// }

	return nil
}

// Get all API keys associated with an account
func (dbm *DBManager) GetAPIKeys(account *Models.Account) ([]Models.APIKey, error) {
	var apiKeys []Models.APIKey
	if err := dbm.DB.Model(account).Association("API_Keys").Find(&apiKeys); err != nil {
		err = fmt.Errorf("Error getting API keys: %v", err)
		glog.Error(err)
		return nil, err
	}

	return apiKeys, nil
}
