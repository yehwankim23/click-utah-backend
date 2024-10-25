# Click Utah Backend

API that manages Click Utah's authentication and database

## API Endpoints

Base URL : `http://click-utah-backend.140.238.11.223.sslip.io`

### /time

Request body

- `none`

Response

- `string` time : Server time
- `stirng` version : API version

### /signup

Request body

- `string` email : Email
- `string` password : Password
- `string` name : Name

Response

- `boolean` error : Error flag
- `string` email : Email
- `string` name : Name
- `string` uid : UID
- `number` count : Click count

### /signin

Request body

- `string` email : Email
- `string` password : Password

Response

- `boolean` error : Error flag
- `string` email : Email
- `string` name : Name
- `string` uid : UID
- `number` count : Click count

### /user

Request body

- `string` uid : UID

Response

- `boolean` error : Error flag
- `string` email : Email
- `string` name : Name
- `string` uid : UID
- `number` count : Click count

### /rename

Request body

- `string` uid : UID
- `string` name : Name

Response

- `boolean` error : Error flag
- `string` email : Email
- `string` name : Name
- `string` uid : UID
- `number` count : Click count

### /click

Request body

- `string` uid : UID

Response

- `boolean` error : Error flag
- `string` email : Email
- `string` name : Name
- `string` uid : UID
- `number` count : Click count

### /leaderboard

Request body

- `none`

Response

- `boolean` error : Error flag
- `array` leaderboard
  - `string` email : Email
  - `string` name : Name
  - `string` uid : UID
  - `number` count : Click count
  - `number` timestamp : Timestamp

## Error Flag

Response when error flag is `true`

- `boolean` error : Error flag
- `string` message : Error message
