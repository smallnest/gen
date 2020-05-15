## Protocol Buffer Code generation
By using the `--protobuf ` flag, `gen` will generate a `.proto` file to describe the models generated from the database tables. It will also add `proto` struct tags to the generated struct.  
You can customize the naming of the fields using the `--proto-fmt` flag. The generated `.proto` will be places in the `--out` directory.   

 `--proto-fmt=snake                               proto name format [snake | camel | lower_camel | none]`
 
```proto

syntax = "proto3";

package main;

message packet {
    enum packet_type_t { 
        PktAlbum = 0;
        PktArtist = 1;
        PktCustomer = 2;
        PktEmployee = 3;
        PktGenre = 4;
        PktInvoiceItem = 6;
        PktInvoice = 5;
        PktMediaType = 7;
        PktPlaylistTrack = 9;
        PktPlaylist = 8;
        PktPurchaseOrder = 11;
        PktTrack = 10;
    }

    packet_type_t id = 2;
    bytes data = 3;
}



message Album { 
    int32 album_id = 1;
    string title = 2;
    int32 artist_id = 3;
}


message Artist { 
    int32 artist_id = 1;
    string name = 2;
}


message Customer { 
    int32 customer_id = 1;
    string first_name = 2;
    string last_name = 3;
    string company = 4;
    string address = 5;
    string city = 6;
    string state = 7;
    string country = 8;
    string postal_code = 9;
    string phone = 10;
    string fax = 11;
    string email = 12;
    int32 support_rep_id = 13;
}


message Employee { 
    int32 employee_id = 1;
    string last_name = 2;
    string first_name = 3;
    string title = 4;
    int32 reports_to = 5;
    uint64 birth_date = 6;
    uint64 hire_date = 7;
    string address = 8;
    string city = 9;
    string state = 10;
    string country = 11;
    string postal_code = 12;
    string phone = 13;
    string fax = 14;
    string email = 15;
}


message Genre { 
    int32 genre_id = 1;
    string name = 2;
}


message InvoiceItem { 
    int32 invoice_line_id = 1;
    int32 invoice_id = 2;
    int32 track_id = 3;
    float unit_price = 4;
    int32 quantity = 5;
}


message Invoice { 
    int32 invoice_id = 1;
    int32 customer_id = 2;
    uint64 invoice_date = 3;
    string billing_address = 4;
    string billing_city = 5;
    string billing_state = 6;
    string billing_country = 7;
    string billing_postal_code = 8;
    float total = 9;
}


message MediaType { 
    int32 media_type_id = 1;
    string name = 2;
}


message PlaylistTrack { 
    int32 playlist_id = 1;
    int32 track_id = 2;
}


message Playlist { 
    int32 playlist_id = 1;
    string name = 2;
}


message PurchaseOrder { 
    int32 id = 1;
    int32 payment_id = 2;
    string full_name = 3;
}


message Track { 
    int32 track_id = 1;
    string name = 2;
    int32 album_id = 3;
    int32 media_type_id = 4;
    int32 genre_id = 5;
    string composer = 6;
    int32 milliseconds = 7;
    int32 bytes = 8;
    float unit_price = 9;
}
```
