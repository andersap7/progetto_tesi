const fs = require('fs');
const { read, modelChaincode, api, mainFunction, getConnection } = require('../utils');

const axios = require('axios');
exports.execute = () => {
    const usage = "usage: node . execute 'walletUser' 'modelName' 'inputPath'";
    return mainFunction(usage, 3, async (args) => {
        const user = args[0]
        const modelName = args[1];
        const input = fs.readFileSync(args[2], {encoding: "base64"});
        const inputBuf = Buffer.from(input);
        const conn = await getConnection(user, "org1", modelChaincode);

        const transaction = await conn.contract.createTransaction("RunModel");
        transaction.setTransient({"input":inputBuf});
        const result = await transaction.submit(modelName);
    
        console.log(result.toString());
        conn.gateway.disconnect();
    });
}

exports.submit = () => {

    const usage = "usage: node . submit 'walletUser' 'modelName' 'ipfsHash' 'inputdef' 'outputdef'";
    return mainFunction(usage, 5, async (args) => {
        const user = args[0];
        const modelName = args[1];
        const hash = args[2];
        const input = read(args[3]);
        const output = read(args[4]);

        const conn = await getConnection(user, "org1", modelChaincode);

        await conn.contract.submitTransaction("SaveModel", modelName, hash, ...Object.values(input), ...Object.values(output));

        conn.gateway.disconnect();
    });
}

exports.authorize = () => {
    const usage = "usage: node . authorize 'walletUser' 'modelName' 'userToAuthorize'";
    return mainFunction(usage, 3, async (args) => {
        const user = args[0];
        const model = args[1];
        const userToAuth = args[2];
        const conn = await getConnection(user, "org1", modelChaincode);

        await conn.contract.submitTransaction('Authorize', model, userToAuth);
        conn.gateway.disconnect();
    });
}

exports.getModel = () => {
    const usage = "usage: node . getModel 'modelName'";

    return mainFunction(usage, 1, async (args) => {
        const id = args[0];
        const res = await axios.get(`${api}models/`, {params: {id}})
        console.log(JSON.stringify(res.data));
    });
}

exports.getAllModels = () => {
    const usage = "usage: node . getAllModels";
    return mainFunction(usage, 0, async () => {
        const res = await axios.get(`${api}models`);
        console.log(JSON.stringify(res.data));
    });
}

exports.getModelsByUser = () => {
    const usage = "usage: node . getModelsByUser 'clientID'";
    return mainFunction(usage, 1, async (args) => {
        const userID = args[0];
        const res = await axios.get(`${api}/models/`, {params:{userID}})
        console.log(JSON.stringify(res.data));
    }
);
}