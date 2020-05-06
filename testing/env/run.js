// only need to unlock first account to send token to all the other accounts
unlock(0);
// send 100 ETH from accnt[0] to all the other accounts
sendAll(100);
// print all account balances, but it won't show immediately due to new block
// needs time
printAll();
// watchPending();
