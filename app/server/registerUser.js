const { Wallets } = require('fabric-network');
const FabricCAServices = require('fabric-ca-client');
const path = require('path');

const { buildCAClient, registerUser, buildCCP} = require('./utils');

const msp = {
    org1:'Org1MSP',
    org2:'Org2MSP',
}

const register = async (user, org) => {
    // const ccp = buildCCP(`./connection/connection-${org}.json`);
    const ccp = buildCCP(path.join(__dirname, 'connection', `connection-${org}.json`, ));
    const caClient = buildCAClient(FabricCAServices, ccp, `ca.${org}.example.com`);

    const walletPath = path.join(__dirname, 'wallet/', org)
    const wallet = await Wallets.newFileSystemWallet(walletPath);

    const enrollmentSecret = await registerUser(caClient, wallet, msp[org], user, `${org}.department1`)

    return enrollmentSecret;
}

module.exports = register;

