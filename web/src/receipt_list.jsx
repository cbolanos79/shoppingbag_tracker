import { Component } from 'react'
import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'

export default class ReceiptList extends Component {
    
    constructor(props) {
        super(props)
    }

  render() {
    return (
        <Row>
            <Col>
                Receipts!
            </Col>
        </Row>
    )
  }
}