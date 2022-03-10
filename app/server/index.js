const express = require('express');
const path = require('path');

const app = express();


// body parser middleware
app.use(express.json());
// app.use(express.urlencoded({ extended: false}));

app.use('/api/users', require('./routes/api/users'));
app.use('/api/models', require('./routes/api/models'));
app.use('/api/tokens', require('./routes/api/token'));


const PORT = process.env.PORT || 5000;

app.listen(PORT, ()=> console.log(`server started on port ${PORT}`));
