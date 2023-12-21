import Alert from 'react-bootstrap/Alert'
import Button from 'react-bootstrap/Button'
import {useState} from 'react'

function AlertDismissible({content, heading="", variant="", show, setShow}) {  
    if (show) {
      return (
        <Alert variant={variant} onClose={() => setShow(false)} dismissible>
          <Alert.Heading>{heading}</Alert.Heading>
          <p>
           {content}
          </p>
        </Alert>
      );
    }
  }

  export default AlertDismissible