export namespace dto {
	
	export class TrafficLightStateDTO {
	    id: string;
	    intersection_id: string;
	    ip_address: string;
	    current_phase_idx: number;
	    transition_state: string;
	    state_string: string;
	    override_state: string;
	    pos_x: number;
	    pos_y: number;
	
	    static createFrom(source: any = {}) {
	        return new TrafficLightStateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.intersection_id = source["intersection_id"];
	        this.ip_address = source["ip_address"];
	        this.current_phase_idx = source["current_phase_idx"];
	        this.transition_state = source["transition_state"];
	        this.state_string = source["state_string"];
	        this.override_state = source["override_state"];
	        this.pos_x = source["pos_x"];
	        this.pos_y = source["pos_y"];
	    }
	}
	export class VehicleStateDTO {
	    id: string;
	    corporation: string;
	    mode: string;
	    speed_kmh: number;
	    heading_deg: number;
	    pos_x: number;
	    pos_y: number;
	    latitude: number;
	    longitude: number;
	    waiting_at_red: boolean;
	    current_edge_id: string;
	
	    static createFrom(source: any = {}) {
	        return new VehicleStateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.corporation = source["corporation"];
	        this.mode = source["mode"];
	        this.speed_kmh = source["speed_kmh"];
	        this.heading_deg = source["heading_deg"];
	        this.pos_x = source["pos_x"];
	        this.pos_y = source["pos_y"];
	        this.latitude = source["latitude"];
	        this.longitude = source["longitude"];
	        this.waiting_at_red = source["waiting_at_red"];
	        this.current_edge_id = source["current_edge_id"];
	    }
	}
	export class SimulationStateDTO {
	    is_running: boolean;
	    is_exporting: boolean;
	    map_loaded: boolean;
	    total_edges: number;
	    total_lanes: number;
	    total_lights: number;
	    mapped_controllers: number;
	    vehicles: VehicleStateDTO[];
	    traffic_lights: TrafficLightStateDTO[];
	    recent_logs: string[];
	    telemetry_url: string;
	
	    static createFrom(source: any = {}) {
	        return new SimulationStateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_running = source["is_running"];
	        this.is_exporting = source["is_exporting"];
	        this.map_loaded = source["map_loaded"];
	        this.total_edges = source["total_edges"];
	        this.total_lanes = source["total_lanes"];
	        this.total_lights = source["total_lights"];
	        this.mapped_controllers = source["mapped_controllers"];
	        this.vehicles = this.convertValues(source["vehicles"], VehicleStateDTO);
	        this.traffic_lights = this.convertValues(source["traffic_lights"], TrafficLightStateDTO);
	        this.recent_logs = source["recent_logs"];
	        this.telemetry_url = source["telemetry_url"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

