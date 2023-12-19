import {useEffect} from 'react'
import {useState} from 'react'

import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'

import { API_URL } from './constants.js'

function ReceiptList() {    
    const [receipts, setReceipts] = useState([])

    useEffect(() => {
        console.log("Initial")

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
            console.log(data.receipts)
            setReceipts(data.receipts)
        }).
        catch(error => {
            console.error(error.toString())
        })

    }, [])

    return (
        <>
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
                                    return <tr>
                                        <td>{item.ID}</td>
                                        <td>{item.Supermarket}</td>
                                        <td>{new Date(Date.parse(item.Date)).toLocaleDateString("es")}</td>
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