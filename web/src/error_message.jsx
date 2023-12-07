import { Component } from 'react';
import Card from 'react-bootstrap/Card'

export default class ErrorMessage extends Component {
    constructor(props) {
        super(props)
        console.log(this.props.data)
    }

    render() {
        return (
            <Card className="bg-danger text-white shadow m-4 text-center">
                <Card.Body>
                  <h3>{this.props.data.message}</h3>
                    <ul>
                        {this.props.data.errors && this.props.data.errors.map((item => {
                            return <li key={item} className="list-unstyled">{item}</li>
                            })
                        )}
                    </ul>  
                </Card.Body>
            </Card>
        )
    }
}