require('dotenv').config();
const Pusher = require('pusher-js');
const nodemailer = require('nodemailer');
const handlebars = require('handlebars');
const fs = require('fs');

const pusherSocket = new Pusher(process.env.PUSHER_APP_KEY, {
  forceTLS: process.env.PUSHER_APP_SECURE === '1' ? true : false,
  cluster: process.env.PUSHER_APP_CLUSTER,
});

const transporter = nodemailer.createTransport({
  service: 'gmail',
  auth: {
    user: process.env.MAILER_EMAIL,
    pass: process.env.MAILER_PASSWORD,
  },
});

const channel = pusherSocket.subscribe('auth');

channel.bind('login', data => {
  fs.readFile('./index.html', { encoding: 'utf-8' }, function(err, html) {
    if (err) {
      throw err;
    }

    const template = handlebars.compile(html);
    const replacements = {
      username: data.user,
      ip: data.ip,
    };

    let mailOptions = {
      from: '"Pusher Tutorial demo" <foo@example.com>',
      to: data.email,
      subject: 'New login into Pusher tutorials demo app',
      html: template(replacements),
    };

    transporter.sendMail(mailOptions, function(error, response) {
      if (error) {
        console.log(error);
        callback(error);
      }
    });
  });
  console.log(data);
});
