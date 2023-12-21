import { Component } from 'react';
import Button from 'react-bootstrap/Button';
import Form from 'react-bootstrap/Form';
import Logout from '../logout.jsx'
import { CLIENT_ID, API_URL } from '../constants.js'

/*
  Component to render a form which sends a post request with form data,
  and call to success function if it was ok or failure if something was wrong
  Success function receives response json from API, and failure receives error message and detail
*/
function upload(event, sucess, failure) {
    event.preventDefault()
}

export default class ReceiptUpload extends Component {
    constructor(props) {
        super(props)
        this.handleSubmit = this.handleSubmit.bind(this)
    }

    /*
      This method sends ajax request to API in order to create a receipt, and handles response
      If response status is >= 400, use failure callback sending error message and data
      If response is successfull, use success callback with received data
    */
    async handleSubmit(event) {
        event.preventDefault()
        this.props.loading()

        const file = document.getElementById("receipt_file").files[0]
        const formData = new FormData()
        formData.append('file', file)

        await fetch(`${API_URL}/receipt`,
            {
                method: "POST",
                mode: "cors",
                body: formData,
                headers: {
                    "Authorization": "Bearer " + sessionStorage.getItem("authtoken")
                }
            }
        ).
        then((response) => {
            if (response.ok) {
                return response.json()
            } else {
                if (response.status == 401) {
                    return (
                        <Logout />
                    )
                } else {
                    response.json().then (data => {
                        this.props.failure({message: data.message, errors: data.errors})
                    })
                }
            }    
        }).
        then(data => {
            this.props.success(data.receipt)
        }).
        catch(error => {
            this.props.failure({message: "Error connecting to API", errors: [error.toString()]})
        })
    }

    render() {
      return (
            <>
                <Form role="form" id="form" onSubmit={this.handleSubmit} encType="multipart/form-data">
                    <Form.Group className="form-group" controlId="receipt_file">
                        <Form.Label>File</Form.Label>
                        <Form.Control type="file" name="file" itemID="file"/>
                    </Form.Group>
                    <Button variant="primary" type="submit">
                        Send
                    </Button>
                </Form>
            </>
        )
    }
}
