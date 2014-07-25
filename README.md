##What is Gondalf?##
Gondalf is a ready to deploy microservice that provides user management, authentication, and role based authorization features out of box. Gondalf is built using [martini](https://github.com/go-martini/martini) and [gorm](https://github.com/jinzhu/gorm), and uses [postgresql](http://www.postgresql.org) as the default database.

##Features:##

###1. User management###
- User creation
- Validating unique username
- Password change on first login
- Encrypted password storage
- Activity logs

###2. Authentication###
- User authentication
- Token-based authentication
- Custom token expiry and renewal times

###3. Authorization###
- Role-based authorization including group permissions


##Why Gondalf?##
Over the course of multiple projects I realized that there are some common features that can be packed into a single microservice and can be used right out of the box. Gondalf is the first piece in that set.


##TODO List##
- [ ] Improve documentation
- [ ] Input timeout values and server port from config file
- [ ] Add SSL support for the end point
- [ ] Dockerize gondalf
- [ ] Provide one click deploy solution
- [ ] Add support for other databases


##Why call it *Gondalf* ?##

Because  <img src="http://www.reactiongifs.com/wp-content/uploads/2013/12/shall-not-pass.gif" width="100px" height="50px"/> and it is Go, so why not both?


 
##Installation instructions:##

- Clone the repository in a local directory

- Database configuration can be set under the the config file named gondalf.config

- Gondalf creates required tables using the configuration provided in the config file. For this the 
initdb flag should be set to true when starting the app.

`$ bash startApp.sh -initdb=true` 