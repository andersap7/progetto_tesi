const { Wallets, Gateway } = require('fabric-network');
const fs = require('fs');

const channelName = "mychannel";

const mnist_model = "QmcyWym6P9KtDTxtf9AFUkWAeWHpTvRRji29XnpaspjXmk";
const cifar_model = "QmRiviyzxLFmy4nAz2ZSjGQPwy7eBmCE6u5w5Cgm8d19cu";

exports.tokenChaincode = "tokens";
exports.modelChaincode = "models";
exports.api = "http://localhost:5000/api/";
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

exports.enrollUser = async (caClient, orgMspId, enrollmentID, enrollmentSecret) => {
    const enrollment = await caClient.enroll({
        enrollmentID: enrollmentID,
        enrollmentSecret: enrollmentSecret,
    });
    const wallet = await Wallets.newFileSystemWallet('./wallet/org1');
    const identity = {
        credentials: {
            certificate: enrollment.certificate,
            privateKey: enrollment.key.toBytes(),
        },
        mspId: orgMspId,
        type: "X.509",
    };
    wallet.put(enrollmentID, identity);
    console.log("User enrolled and identity imported into the wallet");
    return wallet;
}

exports.mainFunction = (usage, numberOfArguments, func) => {
    const printUsage = () => {
        console.log(usage);
        return;
    }
    const run = async (args) => {
        if (args.length != numberOfArguments) {
            printUsage();
            process.exit(1);
        }
        try {
            await func(args);
        } catch(error) {
            console.log(error);
        }
    }
    return {run, printUsage};
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

        return {contract, gateway};
    } catch (e) {
        console.error(e);
    }
}