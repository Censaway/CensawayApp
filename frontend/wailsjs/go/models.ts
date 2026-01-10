export namespace main {
	
	export class Profile {
	    id: string;
	    name: string;
	    key: string;
	    subscription_id: string;
	    created_at: number;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.key = source["key"];
	        this.subscription_id = source["subscription_id"];
	        this.created_at = source["created_at"];
	    }
	}
	export class UserRule {
	    id: string;
	    type: string;
	    value: string;
	    outbound: string;
	
	    static createFrom(source: any = {}) {
	        return new UserRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.value = source["value"];
	        this.outbound = source["outbound"];
	    }
	}
	export class Settings {
	    routing_mode: string;
	    run_mode: string;
	    mixed_port: number;
	    user_rules: UserRule[];
	    ru_domains: string[];
	    auto_connect: boolean;
	    last_profile_id: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.routing_mode = source["routing_mode"];
	        this.run_mode = source["run_mode"];
	        this.mixed_port = source["mixed_port"];
	        this.user_rules = this.convertValues(source["user_rules"], UserRule);
	        this.ru_domains = source["ru_domains"];
	        this.auto_connect = source["auto_connect"];
	        this.last_profile_id = source["last_profile_id"];
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
	export class Subscription {
	    id: string;
	    name: string;
	    url: string;
	    updated_at: number;
	
	    static createFrom(source: any = {}) {
	        return new Subscription(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	        this.updated_at = source["updated_at"];
	    }
	}

}

