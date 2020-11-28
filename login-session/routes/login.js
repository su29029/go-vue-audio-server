var express = require('express');
var mysql = require('mysql');
var router = express.Router();

/* GET users listing. */
router.post('/', function(req, res, next) {
    if (req.session.user == req.body.params.username) {
        console.log("[login]session on:", req.session.user);
        res.status(200).json({
            msg: "session on",
            username: req.session.user.username
        })
        return
    }
    if (req.body.params.username && req.body.params.password) {
        var username = req.body.params.username;
        var password = req.body.params.password;
        var connection = mysql.createConnection({
            host: 'localhost',
            user: 'su',
            password: 'tF#262420228',
            database: 'session_test'
        });
        connection.connect();
        var query = 'select username from users where username = ?';
        var params = [username, password];
        connection.query(query, [username], function(err, result) {
            if (err) {
                console.log("select error:", err.message);
                res.status(500).json({ msg: "error" });
                return;
            }
            // console.log(result);
            if (result.length == 0) {
                res.status(200).json({ msg: "username invalid" });
                return;
            } else {
                query = 'select * from users where username = ? and password = ?';
                connection.query(query, params, function(err, result) {
                    if (err) {
                        console.log("select error:", err.message);
                        res.status(500).json({ msg: "error" });
                        return;
                    }
                    if (result.length == 0) {
                        res.status(200).json({ msg: "password wrong" });
                        return
                    } else {
                        req.session.user = username;
                        res.status(200).json({ msg: "success" });
                    }
                })
            }
        })
    }
});

module.exports = router;