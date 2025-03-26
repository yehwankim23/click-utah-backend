# Click Utah Backend

API that manages Click Utah's authentication and database

## API Endpoints

Base URL : `https://click-utah-backend.yehwan.kim`

### `/time`

Request body

- `none`

Response

- `string` time : Server time
- `stirng` version : API version

### `/signup`

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

### `/signin`

Request body

- `string` email : Email
- `string` password : Password

Response

- `boolean` error : Error flag
- `string` email : Email
- `string` name : Name
- `string` uid : UID
- `number` count : Click count
- `nunmber` token : Token

### `/user`

Request body

- `string` uid : UID
- `nunmber` token : Token

Response

- `boolean` error : Error flag
- `string` email : Email
- `string` name : Name
- `string` uid : UID
- `number` count : Click count

### `/rename`

Request body

- `string` uid : UID
- `string` name : Name
- `nunmber` token : Token

Response

- `boolean` error : Error flag
- `string` email : Email
- `string` name : Name
- `string` uid : UID
- `number` count : Click count

### `/click`

Request body

- `string` uid : UID
- `nunmber` token : Token

Response

- `boolean` error : Error flag
- `string` email : Email
- `string` name : Name
- `string` uid : UID
- `number` count : Click count

### `/leaderboard`

Request body

- `string` uid : UID
- `nunmber` token : Token

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
