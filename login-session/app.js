var createError = require('http-errors');
var express = require('express');
var path = require('path');
var cookieParser = require('cookie-parser');
var session = require('express-session');
var logger = require('morgan');
var cors = require('cors');
var redis = require('redis');
var redisClient = redis.createClient(6379, 'localhost', { auth_pass: 'tF#262420228' });
var redisStore = require('connect-redis')(session);

var index = require('./routes/index');
var register = require('./routes/register');
var login = require('./routes/login');
var logout = require('./routes/logout');
var saveUser = require('./routes/saveuser');
var app = express();

// view engine setup
app.set('views', path.join(__dirname, 'views'));
app.set('view engine', 'ejs');

app.use(logger('dev'));
app.use(express.json());
app.use(express.urlencoded({ extended: false }));
app.use(cookieParser('123456'))
app.use(express.static(path.join(__dirname, 'public')));
app.use(cors({
    origin: "http://localhost:8080",
    credentials: true
}));
app.use(session({
    store: new redisStore({ host: 'localhost', port: 6379, client: redisClient, ttl: 120 }),
    secret: 'login',
    name: 'login',
    resave: false,
    saveUninitialized: false
}))

app.use('/', index);
app.use('/register', register);
app.use('/login', login);
app.use('/logout', logout);
app.use('/saveuser', saveUser);

// catch 404 and forward to error handler
app.use(function(req, res, next) {
    next(createError(404));
});

// error handler
app.use(function(err, req, res, next) {
    // set locals, only providing error in development
    res.locals.message = err.message;
    res.locals.error = req.app.get('env') === 'development' ? err : {};

    // render the error page
    res.status(err.status || 500);
    res.render('error');
});

module.exports = app;