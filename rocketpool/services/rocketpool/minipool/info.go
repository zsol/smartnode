package minipool

import (
    "encoding/hex"
    "errors"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
)


// Minipoool status types
const (
    INITIALIZED = 0
    PRELAUNCH = 1
    STAKING = 2
    LOGGED_OUT = 3
    WITHDRAWN = 4
    CLOSED = 5
    TIMED_OUT = 6
)


// Minipool detail data
type Details struct {
    Address *common.Address
    Status uint8
    StatusType string
    StatusTime time.Time
    StakingDurationId string
    NodeEtherBalanceWei *big.Int
    NodeRplBalanceWei *big.Int
    UserCount *big.Int
    UserDepositCapacityWei *big.Int
    UserDepositTotalWei *big.Int
}


// Minipool status data
type Status struct {
    Status uint8
    StatusBlock *big.Int
    StakingDuration *big.Int
    ValidatorPubkey []byte
}


// Get a minipool's details
// Requires rocketMinipool and rocketPoolToken contracts to be loaded with contract manager
func GetDetails(cm *rocketpool.ContractManager, minipoolAddress *common.Address) (*Details, error) {

    // Minipool details
    details := &Details{
        Address: minipoolAddress,
    }

    // Initialise minipool contract
    minipoolContract, err := cm.NewContract(minipoolAddress, "rocketMinipool")
    if err != nil {
        return nil, errors.New("Error initialising minipool contract: " + err.Error())
    }

    // Data channels
    statusChannel := make(chan uint8)
    statusTimeChannel := make(chan time.Time)
    stakingDurationIdChannel := make(chan string)
    nodeEtherBalanceChannel := make(chan *big.Int)
    nodeRplBalanceChannel := make(chan *big.Int)
    userCountChannel := make(chan *big.Int)
    userDepositCapacityChannel := make(chan *big.Int)
    userDepositTotalChannel := make(chan *big.Int)
    errorChannel := make(chan error)

    // Get status
    go (func() {
        status := new(uint8)
        if err := minipoolContract.Call(nil, status, "getStatus"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool status: " + err.Error())
        } else {
            statusChannel <- *status
        }
    })()

    // Get status time
    go (func() {
        statusChangedTime := new(*big.Int)
        if err := minipoolContract.Call(nil, statusChangedTime, "getStatusChangedTime"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool status changed time: " + err.Error())
        } else {
            statusTimeChannel <- time.Unix((*statusChangedTime).Int64(), 0)
        }
    })()

    // Get staking duration ID
    go (func() {
        stakingDurationId := new(string)
        if err := minipoolContract.Call(nil, stakingDurationId, "getStakingDurationID"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool staking duration ID: " + err.Error())
        } else {
            stakingDurationIdChannel <- *stakingDurationId
        }
    })()

    // Get node ETH balance
    go (func() {
        nodeEtherBalanceWei := new(*big.Int)
        if err := minipoolContract.Call(nil, nodeEtherBalanceWei, "getNodeBalance"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool node ETH balance: " + err.Error())
        } else {
            nodeEtherBalanceChannel <- *nodeEtherBalanceWei
        }
    })()

    // Get node RPL balance
    go (func() {
        nodeRplBalanceWei := new(*big.Int)
        if err := cm.Contracts["rocketPoolToken"].Call(nil, nodeRplBalanceWei, "balanceOf", minipoolAddress); err != nil {
            errorChannel <- errors.New("Error retrieving minipool node RPL balance: " + err.Error())
        } else {
            nodeRplBalanceChannel <- *nodeRplBalanceWei
        }
    })()

    // Get user count
    go (func() {
        userCount := new(*big.Int)
        if err := minipoolContract.Call(nil, userCount, "getUserCount"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool user count: " + err.Error())
        } else {
            userCountChannel <- *userCount
        }
    })()

    // Get user deposit capacity
    go (func() {
        userDepositCapacityWei := new(*big.Int)
        if err := minipoolContract.Call(nil, userDepositCapacityWei, "getUserDepositCapacity"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool user deposit capacity: " + err.Error())
        } else {
            userDepositCapacityChannel <- *userDepositCapacityWei
        }
    })()

    // Get user deposit total
    go (func() {
        userDepositTotalWei := new(*big.Int)
        if err := minipoolContract.Call(nil, userDepositTotalWei, "getUserDepositTotal"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool user deposit total: " + err.Error())
        } else {
            userDepositTotalChannel <- *userDepositTotalWei
        }
    })()

    // Receive minipool data
    for received := 0; received < 8; {
        select {
            case details.Status = <-statusChannel:
                details.StatusType = getStatusType(details.Status)
                received++
            case details.StatusTime = <-statusTimeChannel:
                received++
            case details.StakingDurationId = <-stakingDurationIdChannel:
                received++
            case details.NodeEtherBalanceWei = <-nodeEtherBalanceChannel:
                received++
            case details.NodeRplBalanceWei = <-nodeRplBalanceChannel:
                received++
            case details.UserCount = <-userCountChannel:
                received++
            case details.UserDepositCapacityWei = <-userDepositCapacityChannel:
                received++
            case details.UserDepositTotalWei = <-userDepositTotalChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return details, nil

}


// Get a minipool's status details
// Requires rocketMinipool contract to be loaded with contract manager
func GetStatus(cm *rocketpool.ContractManager, minipoolAddress *common.Address) (*Status, error) {

    // Minipool status
    status := &Status{}

    // Initialise minipool contract
    minipoolContract, err := cm.NewContract(minipoolAddress, "rocketMinipool")
    if err != nil {
        return nil, errors.New("Error initialising minipool contract: " + err.Error())
    }

    // Data channels
    statusChannel := make(chan uint8)
    statusBlockChannel := make(chan *big.Int)
    stakingDurationChannel := make(chan *big.Int)
    validatorPubkeyChannel := make(chan []byte)
    errorChannel := make(chan error)

    // Get status
    go (func() {
        status := new(uint8)
        if err := minipoolContract.Call(nil, status, "getStatus"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool status: " + err.Error())
        } else {
            statusChannel <- *status
        }
    })()

    // Get status block
    go (func() {
        statusBlock := new(*big.Int)
        if err := minipoolContract.Call(nil, statusBlock, "getStatusChangedBlock"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool status changed block: " + err.Error())
        } else {
            statusBlockChannel <- *statusBlock
        }
    })()

    // Get staking duration
    go (func() {
        stakingDuration := new(*big.Int)
        if err := minipoolContract.Call(nil, stakingDuration, "getStakingDuration"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool staking duration: " + err.Error())
        } else {
            stakingDurationChannel <- *stakingDuration
        }
    })()

    // Get validator pubkey
    go (func() {
        depositInput := new([]byte)
        if err := minipoolContract.Call(nil, depositInput, "getDepositInput"); err != nil {
            errorChannel <- errors.New("Error retrieving minipool depositInput data: " + err.Error())
        } else {
            // :TODO: decode using SSZ once library is available
            validatorPubkeyChannel <- (*depositInput)[4:52]
        }
    })()

    // Receive minipool data
    for received := 0; received < 4; {
        select {
            case status.Status = <-statusChannel:
                received++
            case status.StatusBlock = <-statusBlockChannel:
                received++
            case status.StakingDuration = <-stakingDurationChannel:
                received++
            case status.ValidatorPubkey = <-validatorPubkeyChannel:
                received++
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return status, nil

}


// Get a map of all active minipools by validator pubkey
// Requires rocketPool and rocketMinipool contracts to be loaded with contract manager
func GetActiveMinipoolsByValidatorPubkey(cm *rocketpool.ContractManager) (*map[string]common.Address, error) {

    // Get minipool count
    minipoolCountV := new(*big.Int)
    if err := cm.Contracts["rocketPool"].Call(nil, minipoolCountV, "getPoolsCount"); err != nil {
        return nil, errors.New("Error retrieving minipool count: " + err.Error())
    }
    minipoolCount := (*minipoolCountV).Int64()

    // Data channels
    addressChannels := make([]chan *common.Address, minipoolCount)
    validatorPubkeyChannels := make([]chan string, minipoolCount)
    errorChannel := make(chan error)

    // Get minipool addresses
    for mi := int64(0); mi < minipoolCount; mi++ {
        addressChannels[mi] = make(chan *common.Address)
        go (func(mi int64) {
            minipoolAddress := new(common.Address)
            if err := cm.Contracts["rocketPool"].Call(nil, minipoolAddress, "getPoolAt", big.NewInt(mi)); err != nil {
                errorChannel <- errors.New("Error retrieving minipool address: " + err.Error())
            } else {
                addressChannels[mi] <- minipoolAddress
            }
        })(mi)
    }

    // Receive minipool addresses
    minipoolAddresses := make([]*common.Address, minipoolCount)
    for mi := int64(0); mi < minipoolCount; mi++ {
        select {
            case address := <-addressChannels[mi]:
                minipoolAddresses[mi] = address
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Get minipool validator pubkeys
    for mi := int64(0); mi < minipoolCount; mi++ {
        validatorPubkeyChannels[mi] = make(chan string)
        go (func(mi int64) {

            // Initialise minipool contract
            minipoolContract, err := cm.NewContract(minipoolAddresses[mi], "rocketMinipool")
            if err != nil {
                errorChannel <- errors.New("Error initialising minipool contract: " + err.Error())
                return
            }

            // Get validator pubkey
            depositInput := new([]byte)
            if err := minipoolContract.Call(nil, depositInput, "getDepositInput"); err != nil {
                errorChannel <- errors.New("Error retrieving minipool depositInput data: " + err.Error())
            } else {
                // :TODO: decode using SSZ once library is available
                validatorPubkeyChannels[mi] <- hex.EncodeToString((*depositInput)[4:52])
            }

        })(mi)
    }

    // Receive minipool validator pubkeys & build map
    activeMinipools := make(map[string]common.Address)
    for mi := int64(0); mi < minipoolCount; mi++ {
        select {
            case validatorPubkey := <-validatorPubkeyChannels[mi]:
                activeMinipools[validatorPubkey] = *minipoolAddresses[mi];
            case err := <-errorChannel:
                return nil, err
        }
    }

    // Return
    return &activeMinipools, nil

}


// Get the status type by value
func getStatusType(value uint8) string {
    switch value {
        case INITIALIZED: return "initialized"
        case PRELAUNCH: return "pre-launch"
        case STAKING: return "staking"
        case LOGGED_OUT: return "logged out"
        case WITHDRAWN: return "withdrawn"
        case CLOSED: return "closed"
        case TIMED_OUT: return "timed out"
        default: return "unknown"
    }
}

