import React from 'react'
import ReactDOM from 'react-dom/client'
import {
  RouterProvider,
  createBrowserRouter,
} from "react-router-dom"
import Home from './root.jsx'
import ReceiptList from './receipt_list.jsx'
import NewReceipt from './new_receipt.jsx'
import ErrorPage from './error_page.jsx';
import Login from './login.jsx'
import Logout from './logout.jsx'
import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'
import Nav from 'react-bootstrap/Nav';
import Navbar from 'react-bootstrap/Navbar'
import NavDropdown from 'react-bootstrap/NavDropdown';

let logged = false
let authtoken = sessionStorage.getItem("authtoken")

if ((authtoken === undefined) || (authtoken === null)) {
  logged = false
} else {
  logged = true
}

const router = createBrowserRouter([
  {
    path: "/",
    element: logged ? <Home /> : <Login />,
    errorElement: <ErrorPage />,
  },
  {
    path: "/receipt/list",
    element: logged ? <ReceiptList /> : <Login />
  },
  {
    path: "/receipt/upload",
    element: logged ? <NewReceipt /> : <Login />
  },
  {
    path: "/login",
    element: logged ? <Home /> : <Login />,
  },
  {
    path: "/logout",
    element: <Logout />
  }
])

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <>
      { logged && 
      
      <Row>
        <Col md={12}>        
        <Navbar bg="dark" data-bs-theme="dark">
            <Navbar.Brand className="p-1">SBT</Navbar.Brand>
            <Navbar.Toggle aria-controls="basic-navbar-nav" />
            <Navbar.Collapse id="basic-navbar-nav">
              <Nav className="me-auto">
                <Nav.Link href="/">Home</Nav.Link>
                <NavDropdown title="Receipts" id="navbarScrollingDropdown">
                  <NavDropdown.Item href="/receipt/list">List</NavDropdown.Item>
                  <NavDropdown.Item href="/receipt/upload">Upload</NavDropdown.Item>
                </NavDropdown>
                <Nav.Link href="/logout">Logout</Nav.Link>            
              </Nav>
            </Navbar.Collapse>
            <Navbar.Collapse className="justify-content-end">
              <Navbar.Text>
                <img src={sessionStorage.getItem("profile_picture")} 
                     alt={sessionStorage.getItem("name")}
                     className="rounded-circle p-1"
                     height="60px" />
              </Navbar.Text>
            </Navbar.Collapse>
          </Navbar>
        </Col>
      </Row>
      }

      <Row>
        <Col md={12} className="light shadow">
          <RouterProvider router={router} />
        </Col>
      </Row>

  </>
  </React.StrictMode>,
)
