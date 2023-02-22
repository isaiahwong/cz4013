package Client_test;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

public class flight {

    int flight_identifier;
    String source_plc;
    String dest_plc;

    public flight(int flgt_int, String srcp, String destp){
        flight_identifier = flgt_int;
        source_plc = srcp;
        dest_plc = destp;
    }

    public void serialize(){

    }
}
