package htlc

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type HashedTimeLockContract struct {
	contractapi.Contract
}

type Approval struct {
	SenderID              string `json:"SenderID"`
	ReceiverID            string `json:"ReceiverID"`
	TokensToBeTransferred int    `json:"TokensToBeTransferred"`
}

//Creates new tokens, credits them to minter's account and increments total supply of tokens
func (h *HashedTimeLockContract) Mint(ctx contractapi.TransactionContextInterface, noOfTokens int) error {
	//check if invoker is allowed to mint (Attribute-based Access Control)
	clientIsMinter, foundAttribute, err := ctx.GetClientIdentity().GetAttributeValue("minter")

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
func (h *HashedTimeLockContract) Burn(ctx contractapi.TransactionContextInterface, noOfTokens int) error {

	clientIsMinter, foundAttribute, err := ctx.GetClientIdentity().GetAttributeValue("minter")
	//check if invoker is allowed to burn
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
	//TOCHECK: if accountBalance becomes -ve, should it be set to 0 or send back an error?

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

//Create an approval(agreement) between sender and receiver for transfer of tokens
func (h *HashedTimeLockContract) Approve(ctx contractapi.TransactionContextInterface, noOfTokens int, receiverID string) error {

	senderID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("error while getting client ID: %v", err)
	}

	senderAccountBalanceInBytes, err := ctx.GetStub().GetState(senderID)
	if err != nil {
		return fmt.Errorf("error while reading client account from state database: %v", err)
	}

	if senderAccountBalanceInBytes == nil {
		return fmt.Errorf("sender account does not exist")
	}

	senderAccountBalance, err := strconv.Atoi(string(senderAccountBalanceInBytes))
	if err != nil {
		return fmt.Errorf("error while converting account balance from bytes to int: %v", err)
	}

	//check if the current balance is sufficient for transfer
	if noOfTokens > senderAccountBalance {
		return fmt.Errorf("current balance (%d) is not enough for approving transfer of %d tokens", senderAccountBalance, noOfTokens)
	}

	//check if receiver account exists
	//TODO: Should we return an error if the receiver's account does not exist?
	//      For now, we'll just create a new account and add 0 balance to it
	receiverAccountBalanceInBytes, err := ctx.GetStub().GetState(receiverID)
	if err != nil {
		return fmt.Errorf("error while reading client account from state database: %v", err)
	}

	if receiverAccountBalanceInBytes == nil {
		accountBalance := 0
		receiverAccountBalanceInBytes = []byte(strconv.Itoa(accountBalance))
		err = ctx.GetStub().PutState(receiverID, receiverAccountBalanceInBytes)
		if err != nil {
			return fmt.Errorf("error while putting nil balance in receiver's account: %v", err)
		}
	}

	//create/retrieve the approval

	approvalKey, err := ctx.GetStub().CreateCompositeKey("APPROVAL", []string{senderID, receiverID})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for approval: %v", err)
	}
	//check if an approval is in place already
	//TOCHECK: should we return an error if an approval is already in place?
	//		   For now, we'll append noOfTokens to the existing approval
	currentApprovalInBytes, err := ctx.GetStub().GetState(approvalKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve approval from world state: %v", err)
	}

	approval := Approval{
		SenderID:              senderID,
		ReceiverID:            receiverID,
		TokensToBeTransferred: noOfTokens,
	}

	//if a previous approval exists, append noOfTokens to it
	if currentApprovalInBytes != nil {
		err = json.Unmarshal(currentApprovalInBytes, &approval)
		if err != nil {
			return fmt.Errorf("error while unmarshalling existing approval: %v", err)
		}
		approval.TokensToBeTransferred += noOfTokens
	}

	newApprovalInBytes, err := json.Marshal(approval)
	if err != nil {
		return fmt.Errorf("error while marshalling approval: %v", err)
	}

	err = ctx.GetStub().PutState(approvalKey, newApprovalInBytes)
	if err != nil {
		return fmt.Errorf("error while putting approval in ledger: %v", err)
	}

	return nil
}

//Transfers given number of tokens to receiver's account
//(provided that the amount has been approved and there is sufficient balance in sender's account)
func (h *HashedTimeLockContract) Transfer(ctx contractapi.TransactionContextInterface, noOfTokens int, receiverID string) error {

	senderID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("error while getting sender ID: %v", err)
	}

	//check if sender and receiver accounts exist
	senderAccountBalanceInBytes, err := ctx.GetStub().GetState(senderID)
	if senderAccountBalanceInBytes == nil {
		return fmt.Errorf("sender account does not exist")
	}

	receiverAccountBalanceInBytes, err := ctx.GetStub().GetState(receiverID)
	if receiverAccountBalanceInBytes == nil {
		return fmt.Errorf("receiver account does not exist")
	}

	//check if an approval between the sender and receiver exists
	approvalKey, err := ctx.GetStub().CreateCompositeKey("APPROVAL", []string{senderID, receiverID})
	approvalInBytes, err := ctx.GetStub().GetState(approvalKey)
	if approvalInBytes == nil {
		return fmt.Errorf("transfer has not been approved")
	}

	var approval Approval
	err = json.Unmarshal(approvalInBytes, &approval)
	if err != nil {
		return fmt.Errorf("error while unmarshalling approval")
	}

	if noOfTokens > approval.TokensToBeTransferred {
		return fmt.Errorf("number of tokens cannot be higher than those approved")
	}

	senderAccountBalance, err := strconv.Atoi(string(senderAccountBalanceInBytes))
	if err != nil {
		return fmt.Errorf("error while converting sender account balance from bytes to int: %v", err)
	}

	//make sure sender has enough tokens to be able to transfer
	if noOfTokens > senderAccountBalance {
		fmt.Errorf("sender's balance is not enough for transferring %d tokens", noOfTokens)
	}

	//reduce tokens from sender's account
	senderAccountBalance -= noOfTokens

	receiverAccountBalance, err := strconv.Atoi(string(receiverAccountBalanceInBytes))
	if err != nil {
		return fmt.Errorf("error while converting receiver account balance from bytes to int: %v", err)
	}

	//credit tokens to receiver's account
	receiverAccountBalance += noOfTokens

	//put back the new balances in both accounts
	senderAccountBalanceInBytes = []byte(strconv.Itoa(senderAccountBalance))
	err = ctx.GetStub().PutState(senderID, senderAccountBalanceInBytes)
	if err != nil {
		return fmt.Errorf("error while putting back new balance in sender's account: %v", err)
	}

	receiverAccountBalanceInBytes = []byte(strconv.Itoa(receiverAccountBalance))
	err = ctx.GetStub().PutState(receiverID, receiverAccountBalanceInBytes)
	if err != nil {
		return fmt.Errorf("error while putting back new balance in receiver's account: %v", err)
	}

	//decrement the number of tokens in approval, and put it back
	approval.TokensToBeTransferred -= noOfTokens
	approvalInBytes, err = json.Marshal(approval)
	if err != nil {
		return fmt.Errorf("error while marshalling updated approval: %v", err)
	}

	err = ctx.GetStub().PutState(approvalKey, approvalInBytes)
	if err != nil {
		return fmt.Errorf("error while putting approval in ledger: %v", err)
	}

	return nil
}

func (h *HashedTimeLockContract) TransferConditional(ctx contractapi.TransactionContextInterface, receiverID string, noOfTokens int, expiryTimeAsString string, passcode string) error {

	senderID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("error while getting sender ID: %v", err)
	}

	//check if sender and receiver accounts exist
	senderAccountBalanceInBytes, err := ctx.GetStub().GetState(senderID)
	if senderAccountBalanceInBytes == nil {
		return fmt.Errorf("sender account does not exist")
	}

	receiverAccountBalanceInBytes, err := ctx.GetStub().GetState(receiverID)
	if receiverAccountBalanceInBytes == nil {
		return fmt.Errorf("receiver account does not exist")
	}

	//check if expiry time format is correct
	// layout := "Feb 3, 2013 at 7:54pm (PST)"
	// expiryTime, err := time.Parse(layout, expiryTimeAsString)
	// if err != nil {
	// 	return fmt.Errorf("error while parsing time. Expiry time must be in the format \"Feb 3, 2013 at 7:54pm (PST)\"")
	// }

	//generate a checksum from the passcode

	return nil
}
