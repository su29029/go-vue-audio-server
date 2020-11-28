var express = require('express');
var mysql = require('mysql');
var router = express.Router();

/* GET users listing. */
router.post('/', function(req, res, next) {
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
        var params = [username, password];
        var query = 'select username from users where username = ?';
        connection.query(query, [username], function(err, result) {
            if (err) {
                console.log('select error:', err.message);
                res.status(500).json({ msg: "error" });
                return;
            }
            if (result.length != 0) {
                res.status(200).json({ msg: "username is existed" });
                return;
            } else {
                query = 'insert into users (username,password) values (?,?)';
                connection.query(query, params, function(err, result) {
                    if (err) {
                        console.log('insert error:', err.message);
                        res.status(500).json({ msg: "error" });
                        return;
                    }
                    res.status(200).json({ msg: "success" });
                })
            }
        })
    }
})
module.exports = router;