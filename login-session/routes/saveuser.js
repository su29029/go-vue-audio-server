var express = require('express');
var axios = require('axios');
var https = require('https');
var router = express.Router();

router.post('/', function(req, res, next) {
    var username = req.body.params.username;
    console.log("[node]:saveuser");
    axios.post('https://su29029.xyz/saveuser', { params: { "cmd": "saveuser", "userID": req.session.user, "msg": "" } }).then(ret => {
        if (ret.data.msg === "success") {
            res.status(200).json({ msg: "success" });
        } else {
            res.status(400).json({ msg: "falied" });
        }
    }).catch(err => { console.log(err) });
})

module.exports = router;