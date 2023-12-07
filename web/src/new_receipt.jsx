import {useState} from 'react'

import PuffLoader from "react-spinners/ClipLoader";
import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'
import ReceiptUpload from './receipt_upload.jsx'
import ErrorMessage from './error_message.jsx'
import ReceiptDetail from './receipt_detail.jsx'
import Card from 'react-bootstrap/Card'

function NewReceipt() {
  const [processing, setProcessing] = useState(false)
  const [success, setSuccess] = useState(false)
  const [failure, setFailure] = useState(false)
  const [failureData, setFailureData] = useState({})
  const [successData, setSuccessData] = useState({})

  function loadingCallback() {
    setFailure(false)
    setSuccess(false)
    setProcessing(true)
  }
  
  function successCallback(data) {
    setSuccess(true)
    setProcessing(false)
    setSuccessData(data)
  }
  
  function failureCallback(data) {
    setProcessing(false)
    setSuccess(false)
    setFailure(true)
    setFailureData(data)
  }

  return (
    <>
        <Card className="shadow m-4">
          <Card.Body>
              <Card.Title className="bg-light">Upload receipt</Card.Title>
              <hr />
              <ReceiptUpload success={successCallback} failure={failureCallback} loading={loadingCallback} />
            </Card.Body>
        </Card>
        <Row >
          <Col md={12}>
              { processing && 
                <Card className="shadow m-4">
                  <Card.Body>
                    <h3>Processing</h3>
                    <PuffLoader />
                  </Card.Body>
                </Card>
              }
              {
                failure && <ErrorMessage data={failureData} />
              }
              {
                success && 
                <Card className="shadow m-4">
                  <Card.Body>
                    <h3 className="text-center">Receipt information</h3>
                    <hr />
                    <ReceiptDetail data={successData} />
                  </Card.Body>
                </Card>
              }
        </Col>
      </Row>
    </>
  )
}

export default NewReceipt
