import {useEffect} from 'react'
import {useState} from 'react'

import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'
import Modal from 'react-bootstrap/Modal';

import ReceiptDetail from './receipt_detail.jsx'
import { API_URL } from './constants.js'

function ReceiptList() {    

    const [receipts, setReceipts] = useState([])
    const [receipt, setReceipt] = useState({})
    const [showDetail, setShowDetail] = useState(false)

    async function showReceiptDetail(id) {
        await fetch(`${API_URL}/receipts/${id}`,
        {
            method: "GET",
            mode: "cors",
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
                    console.error(response)
                }
            }    
        }).
        then(data => {
            setReceipt(data.receipt)
            setShowDetail(true)
        }).
        catch(error => {
            console.error(error.toString())
        })
    }

    function hideReceiptDetail() {
        setShowDetail(false)
    }

    function ReceiptModal() {
        return (
            <div className="modal show" >
            <Modal show={showDetail}>
              <Modal.Header closeButton onClick={hideReceiptDetail}>
              </Modal.Header>
      
            { receipt  &&
              <Modal.Body>
                <ReceiptDetail data={receipt} />
              </Modal.Body>
            }
            </Modal>
          </div>        
        )
    }

    useEffect(() => {
        fetch(`${API_URL}/receipts`,
        {
            method: "GET",
            mode: "cors",
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
                    console.error(response)
                }
            }    
        }).
        then(data => {
            setReceipts(data.receipts)
        }).
        catch(error => {
            console.error(error.toString())
        })

    }, [])

    return (
        <>
        <ReceiptModal />
        <Row>
            <Col>
                <h1 class="text-center">Receipts</h1>
            </Col>
        </Row>
        <Row>
            <Col md={12}>
                <table class="table">
                    <thead>
                        <tr>
                            <th scope="col">
                                ID
                            </th>
                            <th scope="col">
                                Supermarket
                            </th>
                            <th scope="col">
                                Date
                            </th>
                            <th scope="col">
                                Total
                            </th>
                        </tr>
                    </thead>
                    <tbody>
                            {receipts.map((item) => {
                                    return <tr key={item.ID} onClick={ () => showReceiptDetail(item.ID) }>
                                        <td>{item.ID}</td>
                                        <td>{item.Supermarket}</td>
                                        <td>{new Date(Date.parse(item.Date)).toLocaleDateString(navigator.language)}</td>
                                        <td>{item.Total}</td>
                                    </tr>
                            })}
                        
                    </tbody>
                </table>
            </Col>
        </Row>
        </>
    )
  
}

export default ReceiptList