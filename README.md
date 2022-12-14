
# htlc-erc20-fabric
This is a Go implementation of the Hashed Time Lock Contract for ERC20 tokens in Hyperledger Fabric. We'll use a simplified version of fabric-samples' test network to test the contract, and fabconnect to create identities and initiate transactions and queries.

## Smart Contract functions
Transactions:
* Mint
* Burn
* Approve
* Transfer
* TransferFrom*
* **TransferConditional***
* **Claim***
* **Revert***

Queries:
* BalanceOf*
* ClientAccountBalance*
* ClientAccountID*
* TotalSupply*
* GetHashTimeLock*

Note: The functions marked with an asterisk are incomplete/yet to be implemented

## Hashed Time Lock Design and Flow

**STEP 1: Create the lock** (implemented in TransferConditional()) \
As with the normal transfer (implemented in the Approve and Transfer functions), we'll store the tokens to be transferred as a struct named Approval{senderID, receiverID, tokensToBeTransferred} \

The lock has two inputs: Expiry time and passcode\
We'll store the lock as follows:
* Generate hash of given passcode. 
* Then store the following key-pairs into the database:
   * key1=LOCK_*senderID_receiverID_hash*, value=expiryTime
   * key2=LOCK_*senderID_receiverID_expiryTime*, value=Approval
* This facilitates looking up the expiry time using the passcode, and then look up the Approval using the expiry time


**STEP 2: The recipient will release the lock and claim the tokens** (implemented in Claim()) \
The lock is checked in two steps:
1. Calculate the hash of the input passcode
2. Using this hash, construct key1 and lookup the expiry time
    * If lookup is unsuccessful, the passcode is incorrect. So, the hash lock is incorrect. The receipient cannot claim the agreed-upon tokens until they retrieve the correct passcode from sender
    * If lookup is successful, goes to step 2
3. Is the current time is less than the retrieved expiry time?
    * If Yes: Using the expiryTime in key2, lookup the Approval. Then transfer tokens to recepient. 
    * If No: The time lock has expired. The receipient can no longer claim the agreed-upon tokens


**STEP 3: If tokens remain unclaimed even after the time lock has expired, the sender can transfer the tokens back into their account** (implemented in Revert())
1. Is current time is greater than expiry time?
    * If Yes: Fetch the Approval corresponding to the lock, and transfer tokens back into sender's account
    * If No: The lock is still valid. The sender cannot claim the tokens yet


## SETUP
## test-network
### Prerequisites
To run the test-network, we need the following folders:
* test-network <- from the fabric-samples repo (renamed as fabric-samples-test-network)
* Fabric binaries <- installed from here: https://hyperledger-fabric.readthedocs.io/en/latest/install.html
* config <- installed from the same location as above
Store the binaries and config files in bin/ and config/ folders respectively, inside the cloned fabric-samples-test-network.

### Note on changes made to the original test-network
Since we only need one peer and orderer, we'll use Org1 and OrdererOrg,
I have commented out the peer, fabric-ca,chaincode setup for Org2
Following files have been changed:
 • network.sh
 • scripts/deployCC.sh
I have also removed the extraneous code samples that not relevant to this exercise.

### Network setup steps
Using CLI, cd into the fabric-samples-test-network/test-network and run the following steps:

`./network.sh up -ca`

This creates Org1 {fabric-ca, peer admin id, peer container}, as well as the OrdererOrg {fabric-ca, orderer admin id, orderer container}

> `./network.sh createChannel -c channel1`

This creates channel1, adds peer0.org1 and orderer nodes to the channel and sets peer0.org1 as anchor peer

> `./network.sh deployCC -c channel1 -ccn htlc -ccp ../../chaincode -ccl go`
This packages the chaincode, installs it on peer0.org1, approves the htlc chaincode for Org1 and commits it to channel1

### Firefly Fabconnect
Next, we'll use Fabconnect to create the client identities and to submit transactions and queries.


#### Setting up fabconnect:
The 3 files needed to get it running (docker-compose, connection profile and fabconnect config) are inside the folder htlc/erc20-fabric/ff-test. I've used the setup instructions provided here:
https://github.com/hyperledger/firefly-fabconnect/blob/main/docs/getting-started/test-network.md#configure_fabconnect_testnetwork

From inside the ff-test folder, run:
> `docker-compose up -d`

If successful, go to: http://localhost:3000/ to view the Swagger UI for the fabconnect API

#### Creating the client identities Alice and BondX
In the Swagger UI: 
1. Register: Go to the `POST /identities` section

    For Alice, execute using this request body:

    `{
    "name": "alice",
    "type": "client",
    "maxEnrollments": 0,
    "attributes": {}
    }`

    For BondX, use this request body:\
    `{
    "name": "bondx2",
    "type": "client",
    "maxEnrollments": 0,
    "attributes": {
        "minter": "true:ecert"
    }
    }`

    Note: BondX ID has a special attribute "minter", that provides it the power to invoke Mint (Attribute-Based Access Control). Other client IDs are not allowed to mint Beths.

    For both identities, you'll receive the secret, which is used in the next step.

2. Enroll:
    Go to the `POST /identities/{username}/enroll` section
    For both Alice and BondX, provide the username and secret in the request body, and execute. If successfull, both identities are now enrolled with Org1's CA and ready to use.

###


## TODOs
* Complete the smart contract implementation
* Some of the code used in multiple functions, like getting/setting balances, etc. could be put in separate, internal functions to avoid duplicate code
* Write tests for all functions

## Miscellaneous notes
 • This network could also be implemented by separating the bank admins(minters) and users into separate organizations.

