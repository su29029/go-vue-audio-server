var express = require('express');
var router = express.Router();

router.get('/', function(req, res, next) {

});

router.get('/islogin', function(req, res, next) {
    console.log("req.session.user:", req.session.user);
    if (req.session.user) {
        console.log("session on");
        res.status(200).json({ msg: "session on", user: req.session.user });
    } else {
        console.log("login first");
        res.status(200).json({ msg: "login first" });
    }
});

module.exports = router;