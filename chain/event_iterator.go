// Copyright 2018-2019 Celer Network

package chain

import (
	ethereum "github.com/ethereum/go-ethereum"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// EventIterator is returned from FilterEvent and is used to iterate over the raw logs and unpacked data
type EventIterator struct {
	Event    interface{}           // Event containing the contract specifics and raw log
	Contract *BoundContract        // Generic contract to use for unpacking event data
	Name     string                // Event name to use for unpacking event data
	Logs     chan ethtypes.Log     // Log channel receiving the found contract events
	Sub      ethereum.Subscription // Subscription for errors, completion and termination
	Done     bool                  // Whether the subscription completed delivering logs
	Fail     error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *EventIterator) Next() (ethtypes.Log, bool) {
	// If the iterator failed, stop iterating
	if it.Fail != nil {
		return ethtypes.Log{}, false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.Done {
		select {
		case log := <-it.Logs:
			return log, true

		default:
			return ethtypes.Log{}, false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.Logs:
		return log, true

	case err := <-it.Sub.Err():
		it.Done = true
		it.Fail = err
		return it.Next()
	}
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EventIterator) Close() error {
	it.Sub.Unsubscribe()
	return nil
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *EventIterator) Error() error {
	return it.Fail
}
