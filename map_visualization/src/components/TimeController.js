import React from 'react';

class TimeController extends React.Component {
    
    constructor(props) {
        super(props);

        this.state = {
            min: null,
            max: null,
            simulation_progress: 0,
            initialized: false
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
                debugger;
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
        }
        if (this.state.initialized) {
            return <button type="button" onClick={handle}>Play</button>
        } else {
            return <button type="button" disabled="true">Play</button>
        }
    }

    render() {
        return <div>
            <h1> Hi, it works max = {this.state.max}, min = {this.state.min} </h1>
            <input type="range" min="1" max="100" class="slider" id="myRange" />
            {this.renderPlayButton()}
            <button type="button">Pause</button>
        </div>
    }

}    

export default TimeController;
