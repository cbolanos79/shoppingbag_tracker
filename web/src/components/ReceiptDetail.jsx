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
        if (this.props.data.Items===undefined) {
            return (<></>)
        }

        return (
            <>
                <dl className="row">
                    <dt className="col-md-2 text-start">Supermarket</dt>
                    <dd className="col-md-10 text-start">{this.props.data.Supermarket}</dd>
                    <dt className="col-md-2 text-start">Date</dt>
                    <dd className="col-md-10 text-start">{new Date(Date.parse(this.props.data.Date)).toLocaleDateString(navigator.language)}</dd>
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
                                var class_name = ""
                                if ((this.props.searchItem.length > 0) && (item.Name.search(new RegExp(this.props.searchItem, "i")) >= 0 )) {
                                    class_name = "table-warning"
                                }

                                return <tr key={item.ID} className={class_name}>
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
