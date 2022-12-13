package htlc

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type HashedTimeLockContract struct {
	contractapi.Contract
}

/* No need, since we just need to store the balance
type Account struct {
	AccountID string `json:"AccountID"`
	Balance int `json:"Balance"`
	//Owner string `json:"Owner"`

}
*/

//Creates new tokens, credits them to minter's account and increments total supply of tokens
func (h *HashedTimeLockContract) Mint(ctx contractapi.TransactionContextInterface, noOfTokens int) error {
	//check if invoker is allowed to mint (Attribute-based Access Control)
	clientIsMinter,foundAttribute, err := ctx.GetClientIdentity().GetAttributeValue("minter")

	if !foundAttribute || clientIsMinter == "false" {
		return fmt.Errorf("this ID is not authorized to mint")
	}

	if err != nil {
		return fmt.Errorf("error while getting client id attibute \"minter\"")
	}

	accountID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("error while getting client ID: %v", err)
	}

	
	accountBalanceInBytes, err := ctx.GetStub().GetState(accountID)
	if err != nil {
		return fmt.Errorf("error while reading client account from state database: %v", err)
	}

	accountBalance := 0 //default value, in case the minter's account does not exist yet
	
	if accountBalanceInBytes != nil {
		accountBalance, err = strconv.Atoi(string(accountBalanceInBytes))
		if err != nil {
			return fmt.Errorf("error while converting account balance from bytes to int: %v", err)
		}
	}

	accountBalance += noOfTokens
	accountBalanceInBytes = []byte(strconv.Itoa(accountBalance))
	err = ctx.GetStub().PutState(accountID, accountBalanceInBytes)
	if err != nil {
		return fmt.Errorf("error while putting balance in ledger: %v", err)
	}

	//increment total supply of tokens
	totalSupplyInBytes, err := ctx.GetStub().GetState("TOTAL_SUPPLY")
	if err != nil {
		return fmt.Errorf("error while reading total supply from state database: %v", err)
	}
	totalSupply := 0 
	
	if totalSupplyInBytes != nil {
		totalSupply, err = strconv.Atoi(string(totalSupplyInBytes))
		if err != nil {
			return fmt.Errorf("error while converting total supply from bytes to int: %v", err)
		}
	}
	totalSupply += noOfTokens
	totalSupplyInBytes = []byte(strconv.Itoa(totalSupply))
	err = ctx.GetStub().PutState("TOTAL_SUPPLY", totalSupplyInBytes)
	if err != nil {
		return fmt.Errorf("error while putting total supply in ledger: %v", err)
	}

	return nil
}

//Returns balance tokens of invoking account
func (h *HashedTimeLockContract) BalanceOf(ctx contractapi.TransactionContextInterface) (int, error) {
	accountID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return 0, fmt.Errorf("error while getting client ID: %v", err)
	}

	accountBalanceInBytes, err := ctx.GetStub().GetState(accountID)
	if err != nil {
		return 0, fmt.Errorf("error while reading from state database: %v", err)
	}

	if accountBalanceInBytes == nil {
		return 0, fmt.Errorf("this account does not exist")
	}
	accountBalance, err := strconv.Atoi(string(accountBalanceInBytes))
	if err != nil {
		return 0, fmt.Errorf("error while converting account balance from bytes to int: %v", err)
	}

	return accountBalance, nil
}

//Reduces tokens from minter's account and the total supply
func (h *HashedTimeLockContract) Burn(ctx contractapi.TransactionContextInterface, noOfTokens int) (error) {
	//check if invoker is allowed to burn
	clientIsMinter,foundAttribute, err := ctx.GetClientIdentity().GetAttributeValue("minter")

	if !foundAttribute || clientIsMinter == "false" {
		return fmt.Errorf("this ID is not authorized to burn")
	}

	if err != nil {
		return fmt.Errorf("error while getting client id attibute \"minter\"")
	}

	if noOfTokens <= 0 {
		return fmt.Errorf("number of tokens to be burnt must be greater than 0")
	}

	accountID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("error while getting client ID: %v", err)
	}

	
	accountBalanceInBytes, err := ctx.GetStub().GetState(accountID)
	if err != nil {
		return fmt.Errorf("error while reading client account from state database: %v", err)
	}

	if accountBalanceInBytes == nil {
		return fmt.Errorf("this account does not exist")
	}
	
	accountBalance, err := strconv.Atoi(string(accountBalanceInBytes))
	if err != nil {
		return fmt.Errorf("error while converting account balance from bytes to int: %v", err)
	}
	

	accountBalance -= noOfTokens
	//TODO: if accountBalance becomes -ve, should it be set to 0 or send back an error?

	accountBalanceInBytes = []byte(strconv.Itoa(accountBalance))
	err = ctx.GetStub().PutState(accountID, accountBalanceInBytes)
	if err != nil {
		return fmt.Errorf("error while putting balance in ledger: %v", err)
	}

	//decrement total supply of tokens
	totalSupplyInBytes, err := ctx.GetStub().GetState("TOTAL_SUPPLY")
	if err != nil {
		return fmt.Errorf("error while reading total supply from state database: %v", err)
	}

	if totalSupplyInBytes == nil {
		return fmt.Errorf("no token has been minted yet")
	}
	
	totalSupply, err := strconv.Atoi(string(totalSupplyInBytes))
	if err != nil {
		return fmt.Errorf("error while converting total supply from bytes to int: %v", err)
	}
	
	totalSupply -= noOfTokens
	//TODO: if totalSupply becomes -ve, should it be set to 0 or send back an error?

	totalSupplyInBytes = []byte(strconv.Itoa(totalSupply))
	err = ctx.GetStub().PutState("TOTAL_SUPPLY", totalSupplyInBytes)
	if err != nil {
		return fmt.Errorf("error while putting total supply in ledger: %v", err)
	}

	return nil
}

func (h *HashedTimeLockContract) Transfer(ctx contractapi.TransactionContextInterface, noOfTokens int, receiver string) (error) {
	
}
