import React from 'react';

class TimeController extends React.Component {
    
    constructor(props) {
        super(props);

        this.state = {
            min: null,
            max: null,
            simulation_progress: 0,
            initialized: false,
            simulationRun: false,
        };
    }

    componentDidMount() {
        fetch("/simulation-info/_search", {
            method: 'POST',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
        })
            .then(res => res.json())
            .then(res => {

                if (res.hits.hits.length == 0) {
                    return;
                }

                let max = res.hits.hits[0]._source.simulation_max_ts;
                this.setState({
                    min: 0,
                    max: max,
                    initialized: true
                })
            })
    }

    renderPlayButton() {
        let handle = () => {
            this.props.notifyStartSimulation(this.state.min, this.state.max);
            this.setState({
                simulationRun: true,
            })
        }
        if (this.state.initialized && this.state.simulationRun == false) {
             return <button type="button" onClick={handle}>Play simulation</button>
        } else {
            return <button type="button" disabled="true">Play simulation</button>
        }
    }

    render() {
        return <div>
            {this.renderPlayButton()}
        </div>
    }

}    

export default TimeController;
