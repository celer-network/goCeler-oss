function unlock(idx) {
  // 0 for infinite unlock time
  personal.unlockAccount(eth.accounts[idx], '', 0);
}

function unlockAllAccounts() {
  eth.accounts.forEach(function(account) {
    console.log('Unlocking ' + account + '...');
    personal.unlockAccount(account, '', 0);
  });
}

function sendAll(amt) {
  var amtWei = web3.toWei(amt, 'ether');
  var fromAcnt = eth.accounts[0];
  eth.accounts.slice(1).forEach(function(acnt) {
    eth.sendTransaction({ from: fromAcnt, to: acnt, value: amtWei });
    console.log('Fund [' + acnt + '] ' + amt + ' ETH');
  });
}

function printAll() {
  eth.accounts.forEach(function(acnt) {
    var b = web3.fromWei(eth.getBalance(acnt), 'ether');
    console.log('[' + acnt + ']: ' + b + ' ETH');
  });
}

function watchPending() {
  eth.filter('pending').watch(function() {
    if (miner.hashrate > 0) {
      return;
    }
    console.log('== Pending transactions! Looking for next block...');
    miner.start(8);
  });

  eth.filter('latest').watch(function() {
    if (!pendingTransactions()) {
      console.log('== No transactions left. Stopping miner...');
      miner.stop();
    }
  });
}

function pendingTransactions() {
  return (
    txpool.status.pending ||
    txpool.status.queued ||
    (eth.pendingTransactions && eth.pendingTransactions.length > 0) ||
    eth.getBlock('pending').transactions.length > 0
  );
}

function sendTo(dst, amt) {
  var amtWei = web3.toWei(amt, 'ether');
  eth.sendTransaction({ from: eth.accounts[0], to: dst, value: amtWei });
  console.log('Fund [' + dst + '] ' + amt + ' ETH');
}
