const {tokenChaincode, api, mainFunction, getConnection } = require('../utils');
const axios = require('axios');

exports.approve = () => {
    const usage ="usage: node . approve 'walletUser' 'clientIDToApprove' amount";
    return mainFunction(usage, 3, async () => {

        const user = args[0];
        const userToApprove = args[1];
        const amount = args[2];

        const conn = await getConnection(user, "org1", tokenChaincode);

        await conn.contract.submitTransaction("Approve", userToApprove, amount);

        conn.gateway.disconnect();
    });
}

exports.transferFrom = () => {
    const f = async (args) => {
        const user = args[0];
        const from = args[1];
        const to = args[2];
        const amount = args[3];

        const conn = await getConnection(user, "org1", tokenChaincode);

        await conn.contract.submitTransaction("TransferFrom", from, to, amount);

        conn.gateway.disconnect();
    } 

    const usage = "usage: node . transferFrom 'walletUser' 'from' 'to' amount";
    return mainFunction(usage, 3, f);
}


exports.getAllowance = () => {
    const f = async (args) => {
        const owner = args[0];
        const spender = args[1]
        const res = await axios.get(`${api}tokens/allowance`, {params:{owner, spender}})
        console.log(res.data);
    }

    const usage = "usage: node . getAllowance 'owner' 'spender'";
    return mainFunction(usage, 2, f);
}