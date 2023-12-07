import {useState} from 'react'
import { GoogleOAuthProvider } from '@react-oauth/google';
import { GoogleLogin } from '@react-oauth/google';
import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'
import Card from 'react-bootstrap/Card'
import ErrorMessage from './error_message.jsx'
import { CLIENT_ID, API_URL } from './constants.js'

export default function Login() {
    const [errorData, setErrorData] = useState({})
    const [showError, setShowError] = useState(false)

    /*
        Sends request to API in order to validate credential and get auth token
        Once auth token is fetch, store it into sessionStorage
    */
    function getAuthToken(credential) {
        fetch(`${API_URL}/login/google`, {
            method: "POST",
            headers: {
            "Content-Type": "application/json"
            },
            body: JSON.stringify({credential: credential.credential})
        }).
          then(
            response => {
              if (response.ok) {
                return response.json()
              } else {
                throw "Error connecting to server - status " + response.status
              }
            },
            error => {
              throw new Error(error)
            }).
          then(data => {
            sessionStorage.setItem("authtoken", data.auth_token)
            sessionStorage.setItem("profile_picture", data.picture_url)
            sessionStorage.setItem("name", data.name)

            window.location.href = "/"
          }).
          catch(error => {
            setErrorData({message: "Error doing login", errors: [error.message]})
            setShowError(true)

        })    
    }

    return (
        <>
            <Row>
                <Col className="col-3"></Col>
                <Col className="col-6 text-center">
                    <Card className="shadow m-2">
                    <Row>
                        <Col className="text-start">
                            <img src="https://mdbootstrap.com/img/new/ecommerce/vertical/004.jpg" width="100%" class="rounded float-start" />
                        </Col>
                            <Col>
                                <Card.Title className="p-3">
                                    <h2>Sign in</h2>
                                </Card.Title>
                                <Card.Body>
                                    <hr />
                                    <GoogleOAuthProvider clientId={CLIENT_ID}>
                                        <GoogleLogin
                                            onSuccess={credentialResponse => getAuthToken(credentialResponse) }
                                            onError={() => {
                                                console.log('Login Failed');
                                                setLogin({success: false, message: "Error loggin with google"})
                                            }} />
                                    </GoogleOAuthProvider>
                                </Card.Body>
                            </Col>
                        </Row>
                    </Card>
                </Col>
                <Col className="col-3"></Col>
            </Row>
            <Row>
                <Col className="col-3"></Col>
                <Col className="col-6 text-center">
                    { showError && (
                        <ErrorMessage data={errorData} />
                        )
                    }
                </Col>
            </Row>
        </>
    )
}