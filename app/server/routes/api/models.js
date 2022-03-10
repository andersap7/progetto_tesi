const express = require('express');

const router = express.Router();

const { modelChaincode, getConnection } = require('../../utils')

router.get('/', async (req, res) => {

    try {
        const id = req.query.id;
        const userID = req.query.userID;
        const conn = await getConnection("admin", "org1", modelChaincode);
        if (id) {
            const result = await conn.contract.evaluateTransaction("TestGetModel", id);
            res.json(JSON.parse(result.toString()));
        } else if(userID) {
            const result = await conn.contract.evaluateTransaction("GetModelsByDev", userID);
            console.log(result.toString());
            res.json(JSON.parse(result.toString()));
        } else {
            const result = await conn.contract.evaluateTransaction("GetAllModels");
            
            res.json(JSON.parse(result.toString()));
        }
        conn.gateway.disconnect();

    } catch (error) {
        console.log(error);
        res.status(500).json({error})
    }
})


module.exports = router;