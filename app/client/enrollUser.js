const { Gateway } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');
const { buildCAClient, buildCCP, tokenChaincode: userChaincode, enrollUser } = require('../utils');
const channelName = "mychannel";

async function main() {

    if (process.argv[2] === undefined && process.argv[3] === undefined) {
        console.log('Usage: node enrollUser.js userID secret');
        process.exit(1);
    }

    const userId = process.argv[2];
    const secret = process.argv[3];
    const orgMspId = "Org1MSP";

    try {
        const ccp = await buildCCP('./connection/connection-org1.json');
        const caClient = await buildCAClient(FabricCAServices, ccp, 'ca.org1.example.com');

        const wallet = await enrollUser(caClient, orgMspId, userId, secret);

        const gateway = new Gateway();
        await gateway.connect(ccp, {
            wallet: wallet,
            identity: userId,
            discovery: { enabled: true, asLocalhost: true } ,
        });

        const channel = await gateway.getNetwork(channelName);

        const contract = channel.getContract(userChaincode);

        const result = await contract.submitTransaction("Register", userId);

        console.log(result.toString());
        gateway.disconnect();
    }
    catch (error) {
        console.error(`Error in enrolling user: ${error}`);
        process.exit(1);
    }
}

main()

