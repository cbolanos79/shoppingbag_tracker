import {useState} from 'react'

import PuffLoader from "react-spinners/ClipLoader";
import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'
import ReceiptUpload from './components/ReceiptUpload.jsx'
import ReceiptDetail from './components/ReceiptDetail.jsx'
import Card from 'react-bootstrap/Card'
import AlertDismissible from './components/AlertDismissible.jsx';


function NewReceipt() {
  const [processing, setProcessing] = useState(false)
  const [success, setSuccess] = useState(false)
  const [showAlert, setShowAlert] = useState(false)
  const [successData, setSuccessData] = useState({})
  const [alertContent, setAlertContent] = useState("")
  const [alertHeading, setAlertHeading] = useState("")
  const [alertVariant, setAlertVariant] = useState("danger")

  function loadingCallback() {
    setSuccess(false)
    setProcessing(true)
  }
  
  function successCallback(data) {
    setSuccess(true)
    setProcessing(false)
    setSuccessData(data)
    setAlertHeading("Receipt processed successfully")
    setAlertVariant("success")
    setShowAlert(true)
  }
  
  function failureCallback(data) {
    setProcessing(false)
    setSuccess(false)
    setAlertHeading(data.message)
    setAlertContent(data.errors.map((item) => {
      return <li className="list-unstyled">{item}</li>
    }))
    setAlertVariant("danger")
    setShowAlert(true)
  }

  return (
    <>
        <Row>
          <Col className="text-center">
            { showAlert && (
               <AlertDismissible heading={alertHeading} content={alertContent} variant={alertVariant} show={showAlert} setShow={setShowAlert} />
              )
            }
          </Col>
        </Row>
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
