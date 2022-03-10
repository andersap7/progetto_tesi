const {execute, submit, authorize, getAllModels, getModel, getModelsByUser} = require('./functions/model')
const {approve, transferFrom, getAllowance} = require('./functions/allowance');
const {enroll, buyTokens, getClientID, requestRole, getBalance, getTotalSupply, transfer} = require('./functions/user');
const functions = {
    submit,
    authorize,
    execute,
    approve,
    transferFrom,
    transfer,
    enroll,
    buy:buyTokens,
    getClientID,
    requestRole,
    getAllModels,
    getModelsByUser,
    getModel,
    getBalance,
    getTotalSupply,
    getAllowance,
}

async function main() {
    const fn = process.argv[2];
    const args = process.argv.slice(3)

    if (!fn || !(fn in functions)) {
        Object.values(functions).forEach(element => {
            element().printUsage();
        });
        return;
    }
    try {
        // console.log(args);
        await functions[fn]().run(args);
    } catch (error) {
        console.log(error);
    }
}
main()