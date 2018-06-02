# Synapse API
This repo was created based on a server-side development course project. 

Supports multiple EEG data operations exposed as a simple HTTP R API, with additional features such as API token authorization, HTTP traffic throttle strategy and CORS support for web clients.

* Plumber Documentation - https://www.rplumber.io/docs/routing-and-input.html#endpoints
* EEG Analysis Script - https://github.com/uwccdl/qeeg

To get started, take a look the installation steps and API docs.

## Contents

- [Installation](#installation)
  - [Docker](#docker)
- [Clients](#clients)
- [HTTP API](#http-api)
  - [Authorization](#authorization)
  - [Params](#params)
  - [Endpoints](#get)
- [Authors](#authors)

## Installation

Follow [Prof. Dave Stearns's tutorial on AWS](https://drstearns.github.io/tutorials/deploy2aws/) to set up Docker on Amazon EC2 Linux AMI 2. Once all dependencies have been installed and the Docker service has been started on the web server, run `docker network create appnet` to set up the private network for Docker containers to communicate with each other. Creating a `raw-data` directory in root is required if the upload endpoint is used.

On the local machine, navigate to `servers/gateway` and edit `deploy.sh` to use the Public DNS of the EC2 Instance acquired in the previous step, then execute the script.

```
cd servers/gateway
./deploy.sh
```

### Docker

See [Dockerfile](https://github.com/fredhw/synapse-api/blob/master/servers/qeeg-api/Dockerfile) for image details related to the Plumber R API.


## Clients

- [sarahp39/SynapseSolutions](https://github.com/sarahp39/SynapseSolutions)

## HTTP API

### Authorization

synapse-api supports a simple token-based API authorization.

#### /v1/users
- POST: handles requests for the "users" resource, and allows clients to create new user accounts
    - params: `email`, `userName`, `password`, `passwordConf`, `firstName`, `lastName`

#### /v1/users/me
- GET: get the current user from the session state and respond with that user encoded as JSON object.
- PATCH: update the current user with the JSON in the request body, and respond with the newly updated user, encoded as a JSON object.
    - params: `firstName`, `lastName`

#### /v1/sessions
- POST: handles requests for the "sessions" resource, and allows clients to begin a new session using an existing user's credentials
    - params: `email`, `password`

#### /v1/sessions/mine
- DELETE: handles requests for the "current session" resource, and allows clients to end that session.

### Params

Complete list of currently available params for the qeeg-api microservice. Take a look to each specific endpoint to see which params are supported

- **subject**     `string`      - name of the subject. This will be combined into a single filename along with session.
- **session**     `string`      - name of the session. This will be combined into a single filename along with subject.
- **ch**          `string`      - the specified channel
- **sampling**    `int`         - the sampling rate. Default: 128 Hz.
- **window**      `int`         - the duration (in seconds) of each segment (epoch) used as the bases of the FFT analysis. Default: 2 seconds.
- **sliding**     `float`       - the proportion of each segment that does not overlap with the previous segment. In other words, the proportion of overlap between adjacent segments is (1 - sliding). It needs to be a number between 0 and 1 (not a percentage!). Default: 0.75.

#### POST /v1/upload (gateway)

##### Required headers

- Authorization: the user authentication token
- filename: the full name of the file to be uploaded

Content-Type: `multipart/form-data`

Uploads a selected file into the `raw-data` folder on the server.

#### GET /v1/sumfile
Content-Type: `application/json`

##### Required headers

- Authorization: the user authentication token
- filename: the full name of the file to be fetched

Provides a summary of the EEG input with the following structure:

- **Subject**     `string`      - name of the subject.
- **Version**     `string`      - the current version of the EEG analysis script.
- **Session**     `string`      - name of the session.
- **Sampling**    `number`      - the sampling rate.
- **Windows**     `number`      - the duration (in seconds) of each segment (epoch) used as the bases of the FFT analysis.
- **Duration**    `number`      - the duration of the data.
- **Blinks**      `number`      - the number of blink artifacts in the data.
- **Meta-Blinks** `number`      - the blinks metadata.
- **CH_mean_BAND_power**  `number`  - the mean power of a certain band in a channel. Default: `c(Delta, Theta, Alpha, Low Beta, High Beta, Gamma)`
- **CH_IAF**      `number`      - the individual alpha frequency in the channel.
- **CH_IAF_power**  `number`    - the power of the individual alpha frequency in the channel.
- **Meta_CH_Samples**  `number` - the meta-number of samples for the individual alpha frequency in the channel.
- **Meta_CH_LongestQualitySegment** `number`  - the longest quality segment for the duration recorded in the channel.
- **Meta_CH_SpectralQuality** `number`  - the score for spectral quality in the channel.

Example response:
```json
{
  "Subject": "test",
  "Version": "3.1.3",
  "Session": "rest",
  "Sampling": 128,
  "Window": 2,
  "Sliding": 0.75,
  "Duration": 303.5312,
  "Blinks": "NA",
  "Meta_Blinks": 179,
  "AF3_mean_Delta_power": 14.1959,
  "AF3_mean_Theta_power": 11.0682,
  "AF3_mean_Alpha_power": 8.9671,
  "AF3_mean_Low Beta_power": 8.4959,
  "AF3_mean_Upper Beta_power": 8.1386,
  "AF3_mean_High Beta_power": 7.4463,
  "AF3_mean_Gamma_power": 6.7883,
  "AF3_IAF": 12,
  "AF3_IAF_Power": 8.9041,
  "Meta_AF3_Samples": 36,
  "Meta_AF3_LongestQualitySegment": 303.5312,
  "Meta_AF3_SpectralQuality": 0.3534,
  ...
}
```

#### POST /v1/sumfile
Content-Type: `text/plain`

Downloads the summary data text file.

##### Allowed params
- filename  `string` `required`

#### GET /v1/specfile

##### Required headers

- Authorization: the user authentication token
- filename: the full name of the file to be fetched

Content-Type: `application/json`

Returns the spectrum data in array form.

##### Allowed params

- subject     `string`  `required`
- session     `string`  `required`

```json
[
  [
    "Subject",
    "Channel",
    "0.5Hz",
    "1Hz",
    "1.5Hz",
    "2Hz",
    "2.5Hz",
    "3Hz"
  ],
  [
    "test",
    "AF3",
    "17.0328419883707",
    "15.8056141890707",
    "13.9757509727302",
    "13.453516088988",
    "13.3146532586841",
    "13.0232299791357" 
  ]
]
```

#### POST /v1/specfile
Content-Type: `text/plain`

Downloads the spectra data text file.

##### Allowed params
- filename  `string` `required`

#### GET /v1/cohrfile
Content-Type: `application/json`

##### Required headers

- Authorization: the user authentication token
- filename: the full name of the file to be fetched

Returns the spectrum data in array form.

##### Allowed params

- subject     `string`  `required`
- session     `string`  `required`

#### POST /v1/cohrfile
Content-Type: `text/plain`

Downloads the coherence data text file.

##### Allowed params
- filename  `string` `required`


#### GET /v1/spectrum
Content-Type: `image/png`

Returns a png image of the specified channel data plotted in a spectogram along with blink data

##### Allowed params

- subject     `string`  `required`
- session     `string`  `required`
- ch          `string`  `required`

#### GET /v1/clean
Content-Type: `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`

Downloads the cleaning template excel sheet with the summary data already imported.

##### Allowed params
- filename  `string` `required`

## Authors

- [Frederick Wijaya](https://github.com/fredhw) - Original author and maintainer.
