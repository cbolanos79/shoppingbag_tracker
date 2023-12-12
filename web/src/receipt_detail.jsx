import { Component } from 'react';
import Table from 'react-bootstrap/Table';

/*
    Component to render receipt with main information (supermarket, date, total) and items list formatted as table
*/
export default class ReceiptDetail extends Component {
    constructor(props) {
        super(props)        
    }

    render() {
        return (
            <>
                <dl className="row">
                    <dt className="col-md-2 text-start">Supermarket</dt>
                    <dd className="col-md-10 text-start">{this.props.data.Supermarket}</dd>
                    <dt className="col-md-2 text-start">Date</dt>
                    <dd className="col-md-10 text-start">{this.props.data.Date}</dd>
                    <dt className="col-md-2 text-start">Total</dt>
                    <dd className="col-md-10 text-start">{this.props.data.Total}&nbsp;{this.props.data.Currency}</dd>
                </dl>
                <hr />
                <Table>
                    <thead>
                        <tr>
                            <th>Quantity</th>
                            <th>Name</th>
                            <th>Unit price ({this.props.data.Currency})</th>
                            <th>Price ({this.props.data.Currency})</th>
                        </tr>
                    </thead>
                    <tbody>
                            {this.props.data.Items.map((item => {
                                return <tr key={item.Id}>
                                    <td>{item.Quantity}</td>
                                    <td>{item.Name}</td>
                                    <td>{item.UnitPrice}</td>
                                    <td>{item.Price}</td>
                                </tr>
                            }))}
                    </tbody>
                </Table>

            </>
        )
    }
}