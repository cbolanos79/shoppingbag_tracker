import {
    Navigate
} from "react-router-dom"

export default function Logout() {
    sessionStorage.removeItem("authtoken")
    sessionStorage.removeItem("name")
    sessionStorage.removeItem("profile_picture")
  
    return (
      <Navigate replace to="/login" />
    )
  }