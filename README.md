# aws_learning

## Create rest apis that will be used by user to perform CRUD operation of contact information of users:
### Contact information
- Id
- First name
- Last name
- Address
- Mobile number
- Email address

 

### Components required:
1. Create open spec api that will provide information about apis and their functionality
2. API gateway, Rest apis, Lambda and Dynamodb 
3. Serverless/terraform to deploy above components on cloud
4. APIs operation:
   1. Create user(POST): User shall be able to create user and that to be stored in dynamo
   2. Update user(PATCH): User shall be able to update any user information
   3. Get user(GET):
      - Get user based on id
   4. Get users(GET)
      - Get all users
      - Get all users in alphabetical order of name
      - Get users which contains some name eg. Get users whose name contains 'Rohin'
   5. Delete user(DELETE)
      - Delete user based on id
      - Delete user based on name(either by first name or last name or can also by both)