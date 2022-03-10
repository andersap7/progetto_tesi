const { Wallets, Gateway } = require('fabric-network');
const fs = require('fs');
const path = require('path');

const adminUserId = "admin";
const adminUserPasswd = "adminpw";
const channelName = "mychannel";

exports.tokenChaincode = "tokens";
exports.modelChaincode = "models";

exports.buildCCP = (filepath) => {
    const ccp = this.read(filepath);
    console.log(`Loaded network configuration located at ${filepath}`);
    return ccp;
}

exports.read = (filepath) => {
    const exists = fs.existsSync(filepath);

    if (!exists) {
        throw new Error (`no such file: ${filepath}`);
    }
    const contents = fs.readFileSync(filepath, 'utf-8');

    // build json object from file contents
    return JSON.parse(contents);

}

exports.buildCAClient = (FabricCAServices, ccp, caHostName) => {
    // Creates new CA client for interacting with CA.
    const caInfo = ccp.certificateAuthorities[caHostName];
    const caTLSCACerts = caInfo.tlsCACerts.pem;
    const caClient = new FabricCAServices(caInfo.url, { trustedRoots: caTLSCACerts, verify: false }, caInfo.caName);

    console.log(`Built a CA Client named ${caInfo.caName}`);
    return caClient;
}

exports.enrollAdmin = async (caClient, wallet, orgMspId) => {
    try {
        // check to see if we've already enrolled the admin user.
        const identity = await wallet.get(adminUserId);
        if (identity) {
            console.log("An identity for the admin user already exists in the wallet");
            return
        }

        // enroll the admin user, and import the new identity into the wallet
        const enrollment = await caClient.enroll({ enrollmentID: adminUserId, enrollmentSecret: adminUserPasswd});
        const x509Identity = {
            credentials: {
                certificate: enrollment.certificate,
                privateKey: enrollment.key.toBytes(),
            },
            mspId: orgMspId,
            type: 'X.509',
        };
        await wallet.put(adminUserId, x509Identity);
        console.log("Successfully enrolled admin user and imported it into the wallet");
    }
    catch(error) {
        console.error(`Failed to enroll admin user : ${error}`);
    }
}

exports.registerUser = async (caClient, wallet, orgMspId, userId, affiliation) => {
    const adminIdentity = await wallet.get(adminUserId);
    if (!adminIdentity) {
        console.log('An identity for the admin user does not exist in the wallet');
        console.log('Enroll the admin user before retrying');
        return;
    }
    // build a user object for authenticating with the CA
    const provider = wallet.getProviderRegistry().getProvider(adminIdentity.type);
    const adminUser = await provider.getUserContext(adminIdentity, adminUserId);

    // register the user, enroll the user, and import the new identity into the wallet.
    // if affiliation is specified by client, the affiliation value must be configured in CA
    const secret = await caClient.register({
        affiliation: affiliation,
        enrollmentID: userId,
        role: 'client',
    }, adminUser);

    return secret;
}

exports.getConnection = async (user, org, chaincodeName) => {

    const wallet = await Wallets.newFileSystemWallet(`./wallet/${org}`) 
    const ccp = this.buildCCP(`./connection/connection-${org}.json`);

    const gateway = new Gateway();

    try {
        await gateway.connect(ccp, {
            wallet: wallet,
            identity: user,
            discovery: { enabled: true, asLocalhost: true } // using asLocalhost as this gateway is using a fabric network deployed locally
        });

        const network = await gateway.getNetwork(channelName);

        const contract = network.getContract(chaincodeName);

        const conn = {
            "gateway": gateway,
            "contract": contract,
        }

        return conn;
    } catch (e) {
        console.error(e);
    }
}