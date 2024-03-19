import {useState} from 'react'

import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'
import Modal from 'react-bootstrap/Modal'
import Form from 'react-bootstrap/Form'
import Button from 'react-bootstrap/Button'

import ReceiptDetail from './components/ReceiptDetail.jsx'
import { API_URL } from './constants.js'

function ReceiptList() {    

    const [receipts, setReceipts] = useState([])
    const [receipt, setReceipt] = useState({})
    const [searchItem, setSearchItem] = useState("")
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
                <ReceiptDetail data={receipt} searchItem={searchItem}/>
              </Modal.Body>
            }
            </Modal>
          </div>        
        )
    }

    function search(event) {
        event.preventDefault()
        var params = new URLSearchParams()

        // Set filters as params in query string
        const supermarket = document.getElementById("supermarket").value
        if (supermarket.length > 0) {
            params.append("supermarket", supermarket)
        }
        const min_date = document.getElementById("min_date").value
        if (min_date.length > 0) {
            params.append("min_date", new Date(min_date).toISOString())
        }
        const max_date = document.getElementById("max_date").value
        if (max_date.length > 0) {
            params.append("min_date", new Date(max_date).toISOString())
        }
        const item = document.getElementById("item").value
        if (item.length > 0) {
            params.append("item", item)
        }
        setSearchItem(item)

        const url = `${API_URL}/receipts`

        // Fetch data from API
        fetch(`${url}?${params.toString()}`, {
            method: "GET",
            mode: "cors",
            headers: {
                "Authorization": "Bearer " + sessionStorage.getItem("authtoken"),
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
            if (data != null) {
                setReceipts(data.receipts)
            }
        }).
        catch(error => {
            console.error(error.toString())
        })
    }

    function ReceiptsList() {
        console.log(receipts.length)
        if (receipts.length==0) {
            return (<h3 className="text-center">No items found</h3>)
        }

        if (receipts.length>0) {
            return (
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
            )}
        

    }

    function SearchForm() {
        return (
            <>
                <Form role="form" id="form" onSubmit={search}>
                    <Form.Group className="form-group" >
                        <Form.Label>Supermarket</Form.Label>
                        <Form.Control name="supermarket" id="supermarket"/>
                    </Form.Group>
                    <Form.Group className="form-group" >
                        <Form.Label>Item</Form.Label>
                        <Form.Control name="item" id="item"/>
                    </Form.Group>
                    <Form.Group className="form-group" >
                        <Form.Label>Start date</Form.Label>
                        <Form.Control name="min_date" type="date" id="min_date" />
                    </Form.Group>
                    <Form.Group className="form-group" >
                        <Form.Label>End date</Form.Label>
                        <Form.Control name="max_date" type="date" id="max_date" />
                    </Form.Group>

                    <Button variant="primary" type="submit">
                        Search
                    </Button>
                </Form>
            </>            
        )
    }

    return (
        <>
        <Row>
            <Col>
            </Col>
        
        </Row>
        <ReceiptModal />
        <Row>
            <Col>
                <h1 class="text-center">Receipts</h1>
            </Col>
        </Row>
        <Row>
            <Col md={12}>
                <SearchForm />
            </Col>
        </Row>
        <Row>
            <Col md={12}>
                <ReceiptsList />
            </Col>
        </Row>
        </>
    )
  
}

export default ReceiptList
