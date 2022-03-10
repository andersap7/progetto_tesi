const express = require('express');
const { tokenChaincode, getConnection } = require('../../utils');
const router = express.Router();

router.post('/', async (req, res) => {
    const id = req.body.id;
    const amount = req.body.amount;
    console.log(req.body)

    try {

        const conn = await getConnection("admin", "org2", tokenChaincode);
        // get balance
        const result = await conn.contract.evaluateTransaction("GetBalance")
        const balance = Number(result);
        if (balance < amount) {
            // mint token
            await conn.contract.submitTransaction("Mint", amount-balance);
        }

        // transfer tokens to buyer
        await conn.contract.submitTransaction("Transfer", id, amount);
        res.json({message:`${amount} tokens transferred to ${id}`})

    } catch (error) {
        console.log(error);
        res.json({error});
    }
});

router.get('/balance', async (req, res) => {
    const id = req.query.id;
    console.log(`id: ${id}`)
    const conn = await getConnection("admin", "org2", tokenChaincode);
    try {
        const result = await conn.contract.evaluateTransaction("GetUserBalance", id);
        res.json({balance:Number(result)})
    } catch(error) {
        console.log(error)
        res.json(error)
    }

});

router.get('/totalSupply', async (req, res) => {
    const conn = await getConnection("admin", "org2", tokenChaincode);
    try {
        const result = await conn.contract.evaluateTransaction("TotalSupply");
        res.json({totalSupply:Number(result)})
    } catch(error) {
        console.log(error)
        res.json(error)
    }
})

router.get('/allowance', async (req, res) => {
    const conn = await getConnection("admin", "org2", tokenChaincode);
    try {
        const owner = req.query.owner;
        const spender = req.query.spender;
        const result = await conn.contract.evaluateTransaction("Allowance", owner, spender);
        res.json({allowance:Number(result)})
    } catch(error) {
        console.log(error)
        res.json(error)
    }
})

module.exports = router;