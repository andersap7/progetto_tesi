const { tokenChaincode, getConnection } = require('../../utils');
const register = require('../../registerUser');

const express = require('express');
const router = express.Router();

router.post('/register', async (req, res) => {

    const user = req.body.name;
    try {
        const secret = await register(user, 'org1');
        console.log(secret)
        res.json({secret,message:"user registered"})
    } catch(error) {
        console.log(error);
        res.json({error})
    }
});


router.post('/authorize', async (req, res) => {

    const id = req.body.id;
    const role = req.body.role
    try {
        const conn = await getConnection("admin", "org2", tokenChaincode);

        await conn.contract.submitTransaction('Authorize', id, role);
        // conn.gateway.disconnect();
        res.json({"message": `user authorized`});
    } catch (error) {
        console.log(error);
        res.json(error)
    }
})
module.exports = router;