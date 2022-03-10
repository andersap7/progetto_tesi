const { Wallets } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');

const path = require('path');
const { buildCAClient, buildCCP, enrollAdmin, tokenChaincode, getConnection} = require('./utils');

const mspOrg1 = 'Org1MSP';
const mspOrg2 = 'Org2MSP';


const connectToOrg1CA = async () => {
    console.log('\n--> Enrolling the Org1 CA admin');
    const ccpOrg1 = buildCCP('./connection/connection-org1.json');
    const caOrg1Client = buildCAClient(FabricCAServices, ccpOrg1, 'ca.org1.example.com');

    const walletPathOrg1 = path.join(__dirname, 'wallet/org1');
    const walletOrg1 = await Wallets.newFileSystemWallet(walletPathOrg1);

    await enrollAdmin(caOrg1Client, walletOrg1, mspOrg1);

}


const connectToOrg2CA = async () => {
    console.log('\n--> Enrolling the Org2 CA admin');
    const ccpOrg2 = buildCCP('./connection/connection-org2.json');
    const caOrg2Client = buildCAClient(FabricCAServices, ccpOrg2, 'ca.org2.example.com');

    const walletPathOrg2 = path.join(__dirname, 'wallet/org2');
    const walletOrg2 = await Wallets.newFileSystemWallet(walletPathOrg2);

    await enrollAdmin(caOrg2Client, walletOrg2, mspOrg2);

}

// crea user admin su blockchain
const registerToChaincode = async (org) => {
    const conn = await getConnection("admin", org, tokenChaincode);
    await conn.contract.submitTransaction('Register', "admin");
    const result = await conn.contract.evaluateTransaction("GetClientId");
    const id = result.toString();
    await conn.contract.submitTransaction('Authorize', id, "admin");
    conn.gateway.disconnect();

}

const setPrice = async () => {
    const conn = await getConnection("admin", "org2", tokenChaincode);

    await conn.contract.submitTransaction('SetPrices', 100, 5);
    conn.gateway.disconnect();
}

async function enrollAdmins() {
    await connectToOrg1CA();
    await connectToOrg2CA();
    await registerToChaincode("org2");
    await setPrice();
}

module.exports = enrollAdmins;

enrollAdmins();
// async function main() {

//     if (process.argv[2] === undefined) {
//         console.log('Usage: node enrollAdmin.js Org');
//         process.exit(1);
//     }

//     const org = process.argv[2];

//     try {

//         if (org === 'Org1' || org === 'org1') {
//             await connectToOrg1CA();
//         }
//         else if (org === 'Org2' || org === 'org2') {
//             await connectToOrg2CA();
//         }
//         else {
//             console.log('Usage: node enrollAdmin.js Org');
//             console.log('Org must be Org1 or Org2');
//         }
//     } catch (error) {
//         console.error(`Error in enrolling admin: ${error}`);
//         process.exit(1);
//     }
// }

// main();

