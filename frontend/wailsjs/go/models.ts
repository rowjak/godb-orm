export namespace config {
	
	export class DBConfig {
	    Host: string;
	    Port: number;
	    User: string;
	    Password: string;
	    DBName: string;
	    Driver: string;
	
	    static createFrom(source: any = {}) {
	        return new DBConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Host = source["Host"];
	        this.Port = source["Port"];
	        this.User = source["User"];
	        this.Password = source["Password"];
	        this.DBName = source["DBName"];
	        this.Driver = source["Driver"];
	    }
	}

}

export namespace main {
	
	export class ColumnInfo {
	    name: string;
	    dataType: string;
	    rawType: string;
	    goType: string;
	    isNullable: boolean;
	    isPrimaryKey: boolean;
	    isAutoIncrement: boolean;
	    defaultValue?: string;
	    enumValues?: string[];
	    comment?: string;
	
	    static createFrom(source: any = {}) {
	        return new ColumnInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.dataType = source["dataType"];
	        this.rawType = source["rawType"];
	        this.goType = source["goType"];
	        this.isNullable = source["isNullable"];
	        this.isPrimaryKey = source["isPrimaryKey"];
	        this.isAutoIncrement = source["isAutoIncrement"];
	        this.defaultValue = source["defaultValue"];
	        this.enumValues = source["enumValues"];
	        this.comment = source["comment"];
	    }
	}
	export class ConnectionStatus {
	    connected: boolean;
	    driver: string;
	    host: string;
	    databaseName: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.driver = source["driver"];
	        this.host = source["host"];
	        this.databaseName = source["databaseName"];
	        this.error = source["error"];
	    }
	}

}

