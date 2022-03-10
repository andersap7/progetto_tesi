const axios = require('axios');
const {enrollUser, buildCAClient, buildCCP, tokenChaincode, api, mainFunction, getConnection} = require('../utils')
const FabricCAServices = require('fabric-ca-client');
const { Gateway } = require('fabric-network');

exports.enroll = () => {
    const orgMspId = "Org1MSP";
    const channelName = "mychannel";

    const usage = "usage: node . enroll username";

    return mainFunction(usage, 1, async (args) => {
        const user = args[0];
        const res = await axios.post(`${api}users/register`, {name:user});
        const secret = res.data.secret;
        console.log(secret)
        const ccp = await buildCCP('./connection/connection-org1.json');
        const caClient = await buildCAClient(FabricCAServices, ccp, 'ca.org1.example.com');

        const wallet = await enrollUser(caClient, orgMspId, user, secret);
        const gateway = new Gateway();
        await gateway.connect(ccp, {
            wallet: wallet,
            identity: user,
            discovery: { enabled: true, asLocalhost: true } ,
        });
        const channel = await gateway.getNetwork(channelName);

        const contract = channel.getContract(tokenChaincode);

        const result = await contract.submitTransaction("Register", user);

        console.log(result.toString());
        gateway.disconnect();
    });
}

exports.getClientID = () => {
    const usage = "usage: node . getClientID 'walletUser'"
    return mainFunction(usage, 1, async (args) => {
        const user = args[0];
        const conn = await getConnection(user, "org1", tokenChaincode);
        const result = await conn.contract.evaluateTransaction("GetClientId");
        console.log(result.toString());
    });
}

exports.buyTokens = () => {
    const usage = "usage: node . buy 'clientID' 'amount'";
    return mainFunction(usage, 2, async (args) => {
        const id = args[0];
        const amount = args[1];
        const res = await axios.post(`${api}tokens`, {id, amount})
        console.log(res.data);
    });
}

exports.requestRole = () => {
    const usage = "usage: node . requestRole 'clientID' 'role'";
    return mainFunction(usage, 2, async (args) => {
        const id = args[0];
        const role = args[1];
        const res = await axios.post(`${api}users/authorize`, {id, role})
        console.log(res.data);
    });
}

exports.getTotalSupply = () => {
   const usage = "usage: node . getTotalSupply";
    return mainFunction(usage, 0, async () => {
        const res = await axios.get(`${api}tokens/totalSupply`)
        console.log(res.data);
    });
}
exports.getBalance = () => {
    const usage = "usage: node . getBalance 'clientID'";

    return mainFunction(usage, 1, async (args) => {
        const id = args[0];
        console.log(id);
        const res = await axios.get(`${api}tokens/balance`, {params:{id}})
        console.log(res.data);
    });
}

exports.transfer = () => {
    const usage = "usage: node . transfer 'walletUser' 'to' 'amount'"
    return mainFunction(usage, 3, async (args) => {
        const user = args[0];
        const to = args[1];
        const amount = args[2];

        const conn = await getConnection(user, "org1", tokenChaincode);
        await conn.contract.submitTransaction("Transfer", to, amount);
        conn.gateway.disconnect();
    });
}