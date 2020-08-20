db.createUser(
    {
      user: 'admin',
      pwd: 'putyourstrongpasswordhere',
      roles: ['readWrite', 'dbAdmin']
    }
 )