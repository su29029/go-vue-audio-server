var express = require('express');
var router = express.Router();

router.post('/', function(req, res, next) {
    console.log("[logout]req.session.user:", req.session.user);
    req.session.destroy(err => { console.log(err) })
    res.json({ msg: "success" });
})

module.exports = router;